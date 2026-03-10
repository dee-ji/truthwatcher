package topology

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

var ErrNotFound = errors.New("topology object not found")

type DeviceFilter struct {
	Site     string
	Vendor   string
	Platform string
}

type Service interface {
	Devices(context.Context, DeviceFilter) ([]domain.Device, error)
	Device(context.Context, string) (domain.DeviceDetail, error)
	Links(context.Context, DeviceFilter) ([]domain.Link, error)
	Import(context.Context, domain.TopologySnapshot) error
	Export(context.Context) (domain.TopologySnapshot, error)
	AdjacentDeviceIDs(context.Context, string) ([]string, error)
	BlastRadius(context.Context, string) ([]string, error)
	Paths(context.Context, string, string) ([][]string, error)
}

type Repository interface {
	Import(context.Context, domain.TopologySnapshot) error
	Export(context.Context) (domain.TopologySnapshot, error)
}

type Graph struct {
	adj map[string]map[string]struct{}
}

func buildGraph(snapshot domain.TopologySnapshot) *Graph {
	ifaceToDevice := make(map[string]string, len(snapshot.Interfaces))
	for _, iface := range snapshot.Interfaces {
		ifaceToDevice[iface.ID] = iface.DeviceID
	}
	adj := map[string]map[string]struct{}{}
	for _, link := range snapshot.Links {
		a := ifaceToDevice[link.AInterfaceID]
		z := ifaceToDevice[link.ZInterfaceID]
		if a == "" || z == "" {
			continue
		}
		if adj[a] == nil {
			adj[a] = map[string]struct{}{}
		}
		if adj[z] == nil {
			adj[z] = map[string]struct{}{}
		}
		adj[a][z] = struct{}{}
		adj[z][a] = struct{}{}
	}
	return &Graph{adj: adj}
}

func (g *Graph) Adjacent(id string) []string {
	neigh := g.adj[id]
	out := make([]string, 0, len(neigh))
	for k := range neigh {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

type topologyService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &topologyService{repo: repo}
}

func NewStubService() Service {
	return NewService(NewInMemoryRepository())
}

func (s *topologyService) Devices(ctx context.Context, f DeviceFilter) ([]domain.Device, error) {
	snap, err := s.repo.Export(ctx)
	if err != nil {
		return nil, err
	}
	bySite := toLookup(snap.Sites, func(s domain.Site) string { return s.ID }, func(s domain.Site) string { return s.Name })
	byVendor := toLookup(snap.Vendors, func(v domain.Vendor) string { return v.ID }, func(v domain.Vendor) string { return v.Name })
	byPlatform := toLookup(snap.Platforms, func(p domain.Platform) string { return p.ID }, func(p domain.Platform) string { return p.Name })

	out := make([]domain.Device, 0, len(snap.Devices))
	for _, d := range snap.Devices {
		d.Site = bySite[d.SiteID]
		d.Vendor = byVendor[d.VendorID]
		d.Platform = byPlatform[d.PlatformID]
		if !matches(d, f) {
			continue
		}
		out = append(out, d)
	}
	return out, nil
}

func (s *topologyService) Device(ctx context.Context, id string) (domain.DeviceDetail, error) {
	snap, err := s.repo.Export(ctx)
	if err != nil {
		return domain.DeviceDetail{}, err
	}
	devices, err := s.Devices(ctx, DeviceFilter{})
	if err != nil {
		return domain.DeviceDetail{}, err
	}
	deviceByID := map[string]domain.Device{}
	for _, d := range devices {
		deviceByID[d.ID] = d
	}
	d, ok := deviceByID[id]
	if !ok {
		return domain.DeviceDetail{}, ErrNotFound
	}
	ifaces := []domain.Interface{}
	ifaceDevice := map[string]string{}
	for _, iface := range snap.Interfaces {
		ifaceDevice[iface.ID] = iface.DeviceID
		if iface.DeviceID == id {
			ifaces = append(ifaces, iface)
		}
	}
	links := []domain.Link{}
	for _, link := range snap.Links {
		aDev := ifaceDevice[link.AInterfaceID]
		zDev := ifaceDevice[link.ZInterfaceID]
		if aDev == id || zDev == id {
			link.FromDeviceID, link.ToDeviceID = aDev, zDev
			link.FromDevice = deviceByID[aDev].Hostname
			link.ToDevice = deviceByID[zDev].Hostname
			links = append(links, link)
		}
	}
	adj, _ := s.AdjacentDeviceIDs(ctx, id)
	return domain.DeviceDetail{Device: d, Interfaces: ifaces, AdjacentDeviceIDs: adj, Links: links}, nil
}

func (s *topologyService) Links(ctx context.Context, f DeviceFilter) ([]domain.Link, error) {
	snap, err := s.repo.Export(ctx)
	if err != nil {
		return nil, err
	}
	devices, err := s.Devices(ctx, DeviceFilter{})
	if err != nil {
		return nil, err
	}
	deviceByID := map[string]domain.Device{}
	for _, d := range devices {
		deviceByID[d.ID] = d
	}
	ifaceByID := map[string]domain.Interface{}
	for _, iface := range snap.Interfaces {
		ifaceByID[iface.ID] = iface
	}
	out := []domain.Link{}
	for _, link := range snap.Links {
		a := ifaceByID[link.AInterfaceID]
		z := ifaceByID[link.ZInterfaceID]
		link.FromDeviceID, link.ToDeviceID = a.DeviceID, z.DeviceID
		link.AInterfaceName = a.Name
		link.ZInterfaceName = z.Name
		link.FromDevice = deviceByID[a.DeviceID].Hostname
		link.ToDevice = deviceByID[z.DeviceID].Hostname
		if f != (DeviceFilter{}) {
			if !matches(deviceByID[a.DeviceID], f) && !matches(deviceByID[z.DeviceID], f) {
				continue
			}
		}
		out = append(out, link)
	}
	return out, nil
}

func (s *topologyService) Import(ctx context.Context, snapshot domain.TopologySnapshot) error {
	return s.repo.Import(ctx, snapshot)
}
func (s *topologyService) Export(ctx context.Context) (domain.TopologySnapshot, error) {
	return s.repo.Export(ctx)
}
func (s *topologyService) AdjacentDeviceIDs(ctx context.Context, id string) ([]string, error) {
	snapshot, err := s.repo.Export(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range snapshot.Devices {
		if d.ID == id {
			return buildGraph(snapshot).Adjacent(id), nil
		}
	}
	return nil, ErrNotFound
}

func (s *topologyService) BlastRadius(context.Context, string) ([]string, error) {
	return nil, nil // TODO: future graph analysis extension point.
}
func (s *topologyService) Paths(context.Context, string, string) ([][]string, error) {
	return nil, nil // TODO: future graph analysis extension point.
}

type inMemoryRepository struct {
	mu       sync.RWMutex
	snapshot domain.TopologySnapshot
}

func NewInMemoryRepository() Repository { return &inMemoryRepository{} }

func (r *inMemoryRepository) Import(_ context.Context, snap domain.TopologySnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.snapshot = snap
	return nil
}
func (r *inMemoryRepository) Export(_ context.Context) (domain.TopologySnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.snapshot, nil
}

type postgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) Repository { return &postgresRepository{db: db} }

func (r *postgresRepository) Import(ctx context.Context, s domain.TopologySnapshot) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range []string{"DELETE FROM links", "DELETE FROM interfaces", "DELETE FROM devices", "DELETE FROM platforms", "DELETE FROM vendors", "DELETE FROM sites"} {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	for _, v := range s.Vendors {
		_, err = tx.ExecContext(ctx, "INSERT INTO vendors (id, name) VALUES ($1::uuid, $2)", v.ID, v.Name)
		if err != nil {
			return err
		}
	}
	for _, p := range s.Platforms {
		_, err = tx.ExecContext(ctx, "INSERT INTO platforms (id, vendor_id, name) VALUES ($1::uuid, $2::uuid, $3)", p.ID, p.VendorID, p.Name)
		if err != nil {
			return err
		}
	}
	for _, site := range s.Sites {
		_, err = tx.ExecContext(ctx, "INSERT INTO sites (id, name) VALUES ($1::uuid, $2)", site.ID, site.Name)
		if err != nil {
			return err
		}
	}
	for _, d := range s.Devices {
		_, err = tx.ExecContext(ctx, `INSERT INTO devices (id, hostname, vendor_id, platform_id, site_id) VALUES ($1::uuid,$2,$3::uuid,$4::uuid,$5::uuid)`, d.ID, d.Hostname, nullUUID(d.VendorID), nullUUID(d.PlatformID), nullUUID(d.SiteID))
		if err != nil {
			return err
		}
	}
	for _, iface := range s.Interfaces {
		_, err = tx.ExecContext(ctx, `INSERT INTO interfaces (id, device_id, name) VALUES ($1::uuid,$2::uuid,$3)`, iface.ID, iface.DeviceID, iface.Name)
		if err != nil {
			return err
		}
	}
	for _, l := range s.Links {
		_, err = tx.ExecContext(ctx, `INSERT INTO links (id, a_interface_id, z_interface_id) VALUES ($1::uuid,$2::uuid,$3::uuid)`, l.ID, l.AInterfaceID, l.ZInterfaceID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *postgresRepository) Export(ctx context.Context) (domain.TopologySnapshot, error) {
	var out domain.TopologySnapshot
	rows, err := r.db.QueryContext(ctx, `SELECT id::text, name FROM vendors ORDER BY name`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var v domain.Vendor
		if err := rows.Scan(&v.ID, &v.Name); err != nil {
			return out, err
		}
		out.Vendors = append(out.Vendors, v)
	}
	rows.Close()
	rows, err = r.db.QueryContext(ctx, `SELECT id::text, vendor_id::text, name FROM platforms ORDER BY name`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var p domain.Platform
		if err := rows.Scan(&p.ID, &p.VendorID, &p.Name); err != nil {
			return out, err
		}
		out.Platforms = append(out.Platforms, p)
	}
	rows.Close()
	rows, err = r.db.QueryContext(ctx, `SELECT id::text, name FROM sites ORDER BY name`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var s domain.Site
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return out, err
		}
		out.Sites = append(out.Sites, s)
	}
	rows.Close()
	rows, err = r.db.QueryContext(ctx, `SELECT id::text, hostname, COALESCE(vendor_id::text,''), COALESCE(platform_id::text,''), COALESCE(site_id::text,'') FROM devices ORDER BY hostname`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var d domain.Device
		if err := rows.Scan(&d.ID, &d.Hostname, &d.VendorID, &d.PlatformID, &d.SiteID); err != nil {
			return out, err
		}
		out.Devices = append(out.Devices, d)
	}
	rows.Close()
	rows, err = r.db.QueryContext(ctx, `SELECT id::text, device_id::text, name FROM interfaces ORDER BY device_id, name`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var i domain.Interface
		if err := rows.Scan(&i.ID, &i.DeviceID, &i.Name); err != nil {
			return out, err
		}
		out.Interfaces = append(out.Interfaces, i)
	}
	rows.Close()
	rows, err = r.db.QueryContext(ctx, `SELECT id::text, a_interface_id::text, z_interface_id::text FROM links ORDER BY id`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var l domain.Link
		if err := rows.Scan(&l.ID, &l.AInterfaceID, &l.ZInterfaceID); err != nil {
			return out, err
		}
		out.Links = append(out.Links, l)
	}
	return out, rows.Close()
}

func nullUUID(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func toLookup[T any](in []T, idFn func(T) string, nameFn func(T) string) map[string]string {
	out := map[string]string{}
	for _, item := range in {
		out[idFn(item)] = nameFn(item)
	}
	return out
}

func matches(d domain.Device, f DeviceFilter) bool {
	if f.Site != "" && !strings.EqualFold(d.Site, f.Site) && d.SiteID != f.Site {
		return false
	}
	if f.Vendor != "" && !strings.EqualFold(d.Vendor, f.Vendor) && d.VendorID != f.Vendor {
		return false
	}
	if f.Platform != "" && !strings.EqualFold(d.Platform, f.Platform) && d.PlatformID != f.Platform {
		return false
	}
	return true
}

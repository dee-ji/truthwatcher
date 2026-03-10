import { useNotifications } from '../state/notifications';

export function ToastHost() {
  const { notifications, dismissNotification } = useNotifications();

  return (
    <div className="toast-host" aria-live="polite">
      {notifications.map((notification) => (
        <div className="toast" key={notification.id}>
          <span>{notification.message}</span>
          <button onClick={() => dismissNotification(notification.id)}>Dismiss</button>
        </div>
      ))}
    </div>
  );
}

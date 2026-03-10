import { createContext, useContext, useMemo, useState, type ReactNode } from 'react';

type Notification = {
  id: number;
  message: string;
};

type NotificationContextState = {
  notifications: Notification[];
  pushNotification: (message: string) => void;
  dismissNotification: (id: number) => void;
};

const NotificationContext = createContext<NotificationContextState | null>(null);

export function NotificationProvider({ children }: { children: ReactNode }) {
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const value = useMemo(
    () => ({
      notifications,
      pushNotification: (message: string) => {
        setNotifications((current) => [
          ...current,
          { id: Date.now(), message },
        ]);
      },
      dismissNotification: (id: number) => {
        setNotifications((current) => current.filter((notification) => notification.id !== id));
      },
    }),
    [notifications]
  );

  return <NotificationContext.Provider value={value}>{children}</NotificationContext.Provider>;
}

export function useNotifications() {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotifications must be used within NotificationProvider');
  }
  return context;
}

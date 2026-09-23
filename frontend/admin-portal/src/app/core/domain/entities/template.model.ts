export interface NotificationTemplate {
  id: string;
  key: string;
  name: string;
  channel: 'EMAIL' | 'PUSH' | 'SMS';
  subject: string;
  bodyHtml: string;
  bodyText: string;
  variables: string[];
  createdAt: string;
}

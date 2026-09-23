export interface AdminRole {
  id: string;
  name: string;
  slug: string;
  description: string;
  permissions: string[];
  createdAt: string;
}

export interface CreateRoleDTO {
  name: string;
  slug: string;
  description: string;
  permissions: string[];
}

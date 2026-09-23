import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { RoleRepository } from '../../core/repositories/role.repository';
import { AdminRole, CreateRoleDTO } from '../../core/domain/entities/role.model';
import { ADMIN_API_CONFIG } from '../datasources/admin-api.config';

@Injectable({
  providedIn: 'root'
})
export class RoleRepositoryImpl implements RoleRepository {
  constructor(private http: HttpClient) {}

  listRoles(): Observable<AdminRole[]> {
    return this.http.get<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.roles}`).pipe(
      map(res => {
        const raw = Array.isArray(res) ? res : (res.results || []);
        return raw.map((r: any) => ({
          id: r.id?.toString() || '',
          name: r.name || '',
          slug: r.slug || '',
          description: r.description || '',
          permissions: Array.isArray(r.permissions) ? r.permissions : (typeof r.permissions === 'string' ? JSON.parse(r.permissions) : []),
          createdAt: r.created_at || new Date().toISOString()
        }));
      })
    );
  }

  createRole(role: CreateRoleDTO): Observable<AdminRole> {
    return this.http.post<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.roles}`, {
      name: role.name,
      slug: role.slug,
      description: role.description,
      permissions: role.permissions
    }).pipe(
      map(r => ({
        id: r.id?.toString() || '',
        name: r.name || role.name,
        slug: r.slug || role.slug,
        description: r.description || role.description,
        permissions: r.permissions || role.permissions,
        createdAt: r.created_at || new Date().toISOString()
      }))
    );
  }

  updateRole(id: string, role: CreateRoleDTO): Observable<AdminRole> {
    return this.http.put<any>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.roles}/${id}`, {
      name: role.name,
      slug: role.slug,
      description: role.description,
      permissions: role.permissions
    }).pipe(
      map(r => ({
        id: r.id?.toString() || id,
        name: r.name || role.name,
        slug: r.slug || role.slug,
        description: r.description || role.description,
        permissions: r.permissions || role.permissions,
        createdAt: r.created_at || new Date().toISOString()
      }))
    );
  }

  deleteRole(id: string): Observable<void> {
    return this.http.delete<void>(`${ADMIN_API_CONFIG.baseUrl}${ADMIN_API_CONFIG.endpoints.roles}/${id}`);
  }
}

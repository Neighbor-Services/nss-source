import { Observable } from 'rxjs';
import { AdminRole, CreateRoleDTO } from '../domain/entities/role.model';

export abstract class RoleRepository {
  abstract listRoles(): Observable<AdminRole[]>;
  abstract createRole(role: CreateRoleDTO): Observable<AdminRole>;
  abstract updateRole(id: string, role: CreateRoleDTO): Observable<AdminRole>;
  abstract deleteRole(id: string): Observable<void>;
}

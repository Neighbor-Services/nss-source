import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { RoleRepository } from '../repositories/role.repository';
import { AdminRole, CreateRoleDTO } from '../domain/entities/role.model';

@Injectable({
  providedIn: 'root'
})
export class RoleUseCase {
  constructor(private roleRepo: RoleRepository) {}

  listRoles(): Observable<AdminRole[]> {
    return this.roleRepo.listRoles();
  }

  createRole(role: CreateRoleDTO): Observable<AdminRole> {
    return this.roleRepo.createRole(role);
  }

  updateRole(id: string, role: CreateRoleDTO): Observable<AdminRole> {
    return this.roleRepo.updateRole(id, role);
  }

  deleteRole(id: string): Observable<void> {
    return this.roleRepo.deleteRole(id);
  }
}

import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { UserRepository, ListUsersParams } from '../repositories/user.repository';
import { AdminUser, ImpersonationResult } from '../domain/entities/user.model';

@Injectable({
  providedIn: 'root'
})
export class UserUseCase {
  constructor(private userRepo: UserRepository) {}

  listUsers(params?: ListUsersParams): Observable<{ results: AdminUser[]; count: number }> {
    return this.userRepo.listUsers(params);
  }

  getUserById(id: string): Observable<AdminUser> {
    return this.userRepo.getUserById(id);
  }

  createUser(data: any): Observable<AdminUser> {
    return this.userRepo.createUser(data);
  }

  updateUser(id: string, data: any): Observable<AdminUser> {
    return this.userRepo.updateUser(id, data);
  }

  deleteUser(id: string): Observable<{ success: boolean }> {
    return this.userRepo.deleteUser(id);
  }

  restoreUser(id: string): Observable<{ success: boolean }> {
    return this.userRepo.restoreUser(id);
  }

  impersonateUser(id: string): Observable<ImpersonationResult> {
    return this.userRepo.impersonateUser(id);
  }

  getGDPRUserData(id: string): Observable<any> {
    return this.userRepo.getGDPRUserData(id);
  }

  assignUserRole(userId: string, roleId: string): Observable<{ success: boolean }> {
    return this.userRepo.assignUserRole(userId, roleId);
  }

  listStaffNotes(userId: string): Observable<any[]> {
    return this.userRepo.listStaffNotes(userId);
  }

  createStaffNote(userId: string, data: { content: string; category?: string; is_pinned?: boolean }): Observable<any> {
    return this.userRepo.createStaffNote(userId, data);
  }

  deleteStaffNote(userId: string, noteId: string): Observable<{ status: string }> {
    return this.userRepo.deleteStaffNote(userId, noteId);
  }

  getProviderFunnel(): Observable<any> {
    return this.userRepo.getProviderFunnel();
  }

  getNotificationFeed(): Observable<{ notifications: any[]; count: number; unread_count: number }> {
    return this.userRepo.getNotificationFeed();
  }

  sendDirectMessage(userId: string, title: string, message: string, channel?: string): Observable<{ status: string; sent_at: string }> {
    return this.userRepo.sendDirectMessage(userId, title, message, channel);
  }

  evaluateUserRisk(userId: string): Observable<any> {
    return this.userRepo.evaluateUserRisk(userId);
  }
}

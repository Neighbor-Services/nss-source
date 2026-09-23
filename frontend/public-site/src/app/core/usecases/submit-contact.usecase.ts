import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ContactMessage, ContactResponse } from '../domain/entities/contact.model';
import { ContactRepository } from '../repositories/contact.repository';

@Injectable({
  providedIn: 'root'
})
export class SubmitContactUseCase {
  constructor(private contactRepo: ContactRepository) {}

  execute(message: ContactMessage): Observable<ContactResponse> {
    return this.contactRepo.submitContactMessage(message);
  }
}

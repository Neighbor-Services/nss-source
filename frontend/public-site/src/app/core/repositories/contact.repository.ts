import { Observable } from 'rxjs';
import { ContactMessage, ContactResponse } from '../domain/entities/contact.model';

export abstract class ContactRepository {
  abstract submitContactMessage(message: ContactMessage): Observable<ContactResponse>;
}

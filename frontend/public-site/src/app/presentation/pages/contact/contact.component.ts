import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { SubmitContactUseCase } from '../../../core/usecases/submit-contact.usecase';
import { ContactMessage } from '../../../core/domain/entities/contact.model';

@Component({
  selector: 'app-contact',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './contact.component.html',
  styleUrl: './contact.component.css'
})
export class ContactComponent {
  formData: ContactMessage = {
    firstName: '',
    lastName: '',
    email: '',
    inquiryType: 'general',
    message: ''
  };

  isSubmitting = signal(false);
  alertMessage = signal<string | null>(null);
  isSuccess = signal(false);

  constructor(private submitContactUC: SubmitContactUseCase) {}

  onSubmit(e: Event) {
    e.preventDefault();
    if (!this.formData.firstName || !this.formData.email || !this.formData.message) {
      this.alertMessage.set('Please fill in all required fields.');
      this.isSuccess.set(false);
      return;
    }

    this.isSubmitting.set(true);
    this.alertMessage.set(null);

    this.submitContactUC.execute(this.formData).subscribe({
      next: (res) => {
        this.isSubmitting.set(false);
        if (res.status === 'success') {
          this.isSuccess.set(true);
          this.alertMessage.set(res.message);
          this.formData = {
            firstName: '',
            lastName: '',
            email: '',
            inquiryType: 'general',
            message: ''
          };
        } else {
          this.isSuccess.set(false);
          this.alertMessage.set(res.message);
        }
      },
      error: () => {
        this.isSubmitting.set(false);
        this.isSuccess.set(false);
        this.alertMessage.set('An unexpected error occurred. Please try again or call support.');
      }
    });
  }
}

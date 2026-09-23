import { Component, AfterViewInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { SubmitResolutionUseCase } from '../../../core/usecases/submit-resolution.usecase';

interface ResolutionState {
  role: string;
  roleLabel: string;
  cat: string;
  catLabel: string;
  hint: string;
  ref: string;
  other: string;
  date: string;
  desc: string;
  outcome: string[];
}

@Component({
  selector: 'app-resolution',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './resolution.component.html',
  styleUrl: './resolution.component.css'
})
export class ResolutionComponent implements AfterViewInit {
  step = 0;
  isSubmitting = false;
  submitted = false;
  copied = false;

  outcomes = ['Refund', 'Partial refund', 'Redo the work', 'Cancel booking', 'Apology', 'Other'];

  state: ResolutionState = {
    role: '', roleLabel: '',
    cat: '', catLabel: '', hint: '',
    ref: '', other: '', date: '', desc: '',
    outcome: []
  };

  // Keep original submit capability
  alertMessage = signal<string | null>(null);
  caseId = signal<string | null>(null);
  isSuccess = signal(false);

  constructor(private submitResolutionUC: SubmitResolutionUseCase) {}

  ngAfterViewInit() {
    const io = new IntersectionObserver(entries => {
      entries.forEach(e => {
        if (e.isIntersecting) { e.target.classList.add('in'); io.unobserve(e.target); }
      });
    }, { threshold: 0.14 });
    document.querySelectorAll('.rv').forEach(el => io.observe(el));
  }

  setRole(role: string, label: string) {
    this.state.role = role;
    this.state.roleLabel = label;
  }

  setCat(cat: string, label: string, hint: string) {
    this.state.cat = cat;
    this.state.catLabel = label;
    this.state.hint = hint;
  }

  toggleOutcome(out: string) {
    const idx = this.state.outcome.indexOf(out);
    if (idx > -1) this.state.outcome.splice(idx, 1);
    else this.state.outcome.push(out);
  }

  canAdvance(): boolean {
    if (this.step === 0) return !!this.state.role;
    if (this.step === 1) return !!this.state.cat;
    if (this.step === 2) return this.state.desc.trim().length > 3;
    return true;
  }

  goNext() {
    if (this.step === 3) {
      this.resetForm();
      return;
    }
    if (this.canAdvance()) this.step++;
  }

  goBack() {
    if (this.step > 0) this.step--;
  }

  summaryText(): string {
    const lines = [
      `Role: ${this.state.roleLabel}`,
      `Issue: ${this.state.catLabel}`,
      this.state.ref ? `Booking ref: ${this.state.ref}` : '',
      this.state.other ? `Other neighbor: ${this.state.other}` : '',
      this.state.date ? `Date: ${this.state.date}` : '',
      ``,
      `Description:\n${this.state.desc}`,
      this.state.outcome.length ? `\nDesired outcome: ${this.state.outcome.join(', ')}` : '',
    ].filter(l => l !== undefined);
    return lines.join('\n');
  }

  copyReport() {
    navigator.clipboard.writeText(this.summaryText()).then(() => {
      this.copied = true;
      setTimeout(() => this.copied = false, 2500);
    });
  }

  submitReport() {
    if (!this.state.desc.trim()) return;
    this.isSubmitting = true;

    this.submitResolutionUC.execute({
      role: this.state.role as any,
      issueType: this.state.cat as any,
      bookingRef: this.state.ref,
      otherNeighbor: this.state.other,
      description: this.state.desc,
      expectedOutcome: this.state.outcome.join(', ')
    }).subscribe({
      next: (res) => {
        this.isSubmitting = false;
        this.submitted = true;
        this.caseId.set(res.caseId || null);
      },
      error: () => {
        this.isSubmitting = false;
        this.submitted = true; // Mark as submitted locally even if API fails
      }
    });
  }

  resetForm() {
    this.step = 0;
    this.submitted = false;
    this.copied = false;
    this.state = {
      role: '', roleLabel: '',
      cat: '', catLabel: '', hint: '',
      ref: '', other: '', date: '', desc: '',
      outcome: []
    };
  }
}

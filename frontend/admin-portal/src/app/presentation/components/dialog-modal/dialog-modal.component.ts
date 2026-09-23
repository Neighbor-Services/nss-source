import { Component, inject, HostListener, effect, ViewChild, ElementRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DialogService } from '../../../core/services/dialog.service';

@Component({
  selector: 'app-dialog-modal',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './dialog-modal.component.html',
  styleUrl: './dialog-modal.component.css'
})
export class DialogModalComponent {
  readonly dialogService = inject(DialogService);
  promptInputText: string = '';

  @ViewChild('promptInput') promptInputElement?: ElementRef<HTMLInputElement>;

  constructor() {
    effect(() => {
      const active = this.dialogService.activeDialog();
      if (active) {
        this.promptInputText = active.promptValue || '';
        if (active.type === 'prompt') {
          setTimeout(() => {
            this.promptInputElement?.nativeElement?.focus();
            this.promptInputElement?.nativeElement?.select();
          }, 50);
        }
      }
    });
  }

  @HostListener('window:keydown.escape')
  onEscape() {
    if (this.dialogService.activeDialog()) {
      this.cancel();
    }
  }

  @HostListener('window:keydown.enter', ['$event'])
  onEnter(event: Event) {
    if (this.dialogService.activeDialog()) {
      event.preventDefault();
      this.confirm();
    }
  }

  confirm() {
    const dialog = this.dialogService.activeDialog();
    if (!dialog) return;

    if (dialog.type === 'prompt') {
      this.dialogService.close(this.promptInputText.trim());
    } else if (dialog.type === 'alert' || dialog.type === 'info' || dialog.type === 'success') {
      this.dialogService.close(true);
    } else {
      this.dialogService.close(true);
    }
  }

  cancel() {
    const dialog = this.dialogService.activeDialog();
    if (!dialog) return;

    if (dialog.type === 'prompt') {
      this.dialogService.close(null);
    } else {
      this.dialogService.close(false);
    }
  }
}

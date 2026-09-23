import { Injectable, signal } from '@angular/core';

export type DialogType = 'confirm' | 'alert' | 'prompt' | 'danger' | 'warning' | 'info' | 'success';

export interface DialogConfig {
  id?: string;
  type?: DialogType;
  title: string;
  message: string;
  subMessage?: string;
  confirmText?: string;
  cancelText?: string;
  confirmButtonClass?: string;
  isDanger?: boolean;
  promptPlaceholder?: string;
  promptValue?: string;
  resolve?: (value: any) => void;
}

@Injectable({
  providedIn: 'root'
})
export class DialogService {
  readonly activeDialog = signal<DialogConfig | null>(null);

  /**
   * Prompts user with a confirmation dialog (OK / Cancel)
   * Supports either config object or (title, message, confirmText)
   */
  confirm(
    configOrTitle: {
      title?: string;
      message: string;
      confirmText?: string;
      cancelText?: string;
      isDanger?: boolean;
    } | string,
    message?: string,
    confirmText?: string
  ): Promise<boolean> {
    return new Promise<boolean>((resolve) => {
      let cfg: DialogConfig;
      if (typeof configOrTitle === 'string') {
        cfg = {
          type: 'confirm',
          title: configOrTitle,
          message: message || '',
          confirmText: confirmText || 'Confirm',
          cancelText: 'Cancel',
          resolve
        };
      } else {
        cfg = {
          type: configOrTitle.isDanger ? 'danger' : 'confirm',
          title: configOrTitle.title || 'Confirm Action',
          message: configOrTitle.message,
          confirmText: configOrTitle.confirmText || (configOrTitle.isDanger ? 'Delete' : 'Confirm'),
          cancelText: configOrTitle.cancelText || 'Cancel',
          isDanger: configOrTitle.isDanger,
          resolve
        };
      }
      this.activeDialog.set(cfg);
    });
  }

  /**
   * Prompts user with a destructive danger confirmation dialog
   */
  dangerConfirm(title: string, message: string, confirmText: string = 'Delete'): Promise<boolean> {
    return this.confirm({
      title,
      message,
      confirmText,
      cancelText: 'Cancel',
      isDanger: true
    });
  }

  /**
   * Shows an informational alert dialog with an "OK" button
   */
  alert(config: {
    title?: string;
    message: string;
    type?: DialogType;
    confirmText?: string;
  } | string): Promise<void> {
    return new Promise<void>((resolve) => {
      let cfg: DialogConfig;
      if (typeof config === 'string') {
        cfg = {
          type: 'info',
          title: 'Notification',
          message: config,
          confirmText: 'Got It',
          resolve: () => resolve()
        };
      } else {
        cfg = {
          type: config.type || 'info',
          title: config.title || 'Notification',
          message: config.message,
          confirmText: config.confirmText || 'Got It',
          resolve: () => resolve()
        };
      }
      this.activeDialog.set(cfg);
    });
  }

  /**
   * Prompts user for a text input response
   */
  prompt(config: {
    title: string;
    message: string;
    placeholder?: string;
    defaultValue?: string;
    confirmText?: string;
    cancelText?: string;
  }): Promise<string | null> {
    return new Promise<string | null>((resolve) => {
      const cfg: DialogConfig = {
        type: 'prompt',
        title: config.title,
        message: config.message,
        promptPlaceholder: config.placeholder || 'Enter value...',
        promptValue: config.defaultValue || '',
        confirmText: config.confirmText || 'Submit',
        cancelText: config.cancelText || 'Cancel',
        resolve
      };
      this.activeDialog.set(cfg);
    });
  }

  close(result: any = null) {
    const dialog = this.activeDialog();
    if (dialog && dialog.resolve) {
      dialog.resolve(result);
    }
    this.activeDialog.set(null);
  }
}

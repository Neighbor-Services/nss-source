import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { DialogModalComponent } from './presentation/components/dialog-modal/dialog-modal.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, DialogModalComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})
export class AppComponent {
  title = 'Neighbor Service Admin Portal';
}

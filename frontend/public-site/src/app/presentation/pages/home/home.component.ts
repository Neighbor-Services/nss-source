import { Component, OnInit, AfterViewInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { GetCMSContentUseCase } from '../../../core/usecases/get-cms-content.usecase';
import { CMSContent } from '../../../core/domain/entities/cms.model';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './home.component.html',
  styleUrl: './home.component.css'
})
export class HomeComponent implements OnInit, AfterViewInit {
  cms = signal<CMSContent | null>(null);
  searchQuery = '';
  openFaqId = signal<string | null>(null);
  activePhoneView = signal<'SEEKER' | 'PROVIDER'>('SEEKER');

  constructor(
    private getCMSContentUC: GetCMSContentUseCase,
    private router: Router,
    private sanitizer: DomSanitizer
  ) {}

  ngOnInit() {
    this.getCMSContentUC.execute().subscribe(content => {
      this.cms.set(content);
    });
  }

  ngAfterViewInit() {
    // Scroll reveal
    const io = new IntersectionObserver(entries => {
      entries.forEach(e => {
        if (e.isIntersecting) {
          e.target.classList.add('in');
          io.unobserve(e.target);
        }
      });
    }, { threshold: 0.16 });
    document.querySelectorAll('.rv').forEach(el => io.observe(el));

    // Security code match animation
    const stage = document.getElementById('codeStage');
    if (stage) {
      const io2 = new IntersectionObserver(entries => {
        entries.forEach(e => {
          if (e.isIntersecting) {
            setTimeout(() => {
              stage.classList.add('matched');
              const s = stage.querySelector('[data-status]');
              if (s) s.textContent = 'Service start confirmed ✓';
              const l = stage.querySelector('[data-link-label]');
              if (l) l.textContent = 'Codes matched ✓';
            }, 900);
            io2.unobserve(stage);
          }
        });
      }, { threshold: 0.5 });
      io2.observe(stage);
    }

    // Hero typewriter
    this.initTypewriter();
  }

  private initTypewriter() {
    const qs = [
      'My kitchen sink is leaking and I need someone today…',
      'Photographer for a birthday party next Saturday…',
      'Weekly lawn care, medium yard, Bedford…',
      'Math tutor for my 7th grader, twice a week…'
    ];
    const inp = document.getElementById('heroInput') as HTMLInputElement;
    if (!inp) return;
    let qi = 0, ci = 0, del = false;
    const prefersReduced = matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReduced) { inp.placeholder = qs[0]; return; }

    (function type() {
      const q = qs[qi];
      if (document.activeElement === inp && inp.value) { setTimeout(type, 1200); return; }
      inp.placeholder = q.slice(0, ci);
      if (!del) {
        ci++;
        if (ci > q.length) { del = true; setTimeout(type, 1700); return; }
      } else {
        ci--;
        if (ci === 0) { del = false; qi = (qi + 1) % qs.length; }
      }
      setTimeout(type, del ? 16 : 42);
    })();
  }

  onSearch(e: Event) {
    e.preventDefault();
    const query = this.searchQuery.trim();
    const inp = document.getElementById('heroInput') as HTMLInputElement;
    const v = (query || inp?.placeholder || '').toLowerCase();
    let cat = 'Handyman';
    if (v.includes('sink') || v.includes('leak') || v.includes('plumb')) cat = 'Plumbing';
    else if (v.includes('photo')) cat = 'Photography';
    else if (v.includes('lawn') || v.includes('yard')) cat = 'Landscaping';
    else if (v.includes('tutor') || v.includes('math')) cat = 'Tutoring';
    else if (v.includes('clean')) cat = 'House Cleaning';
    else if (v.includes('mov')) cat = 'Moving';
    else if (v.includes('electr') || v.includes('outlet') || v.includes('fan')) cat = 'Electrical';

    const chip = document.getElementById('demoChip');
    if (chip) {
      chip.innerHTML = `✓ Understood — <b>${cat}</b>&nbsp;· near you · screened matches in the app`;
      chip.classList.add('show');
    }
    setTimeout(() => {
      const dl = document.querySelector('#download');
      if (dl) dl.scrollIntoView({ behavior: 'smooth' });
    }, 1400);
  }

  toggleFaq(id: string) {
    this.openFaqId.update(current => (current === id ? null : id));
  }

  getCategoryIcon(name: string): SafeHtml {
    const lower = name.toLowerCase();
    let svg: string;
    if (lower.includes('plumb'))
      svg = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>';
    else if (lower.includes('clean'))
      svg = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>';
    else if (lower.includes('electr'))
      svg = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>';
    else if (lower.includes('handy') || lower.includes('repair'))
      svg = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>';
    else if (lower.includes('lawn') || lower.includes('garden') || lower.includes('land'))
      svg = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22V12M12 12C12 12 7 9 4 5c4 1 7 4 8 7M12 12c0 0 5-3 8-7-4 1-7 4-8 7"/></svg>';
    else if (lower.includes('paint'))
      svg = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M19 3H5a2 2 0 0 0-2 2v4a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2z"/><path d="M17 11v6a2 2 0 0 1-2 2H9a2 2 0 0 1-2-2v-6"/></svg>';
    else if (lower.includes('mov') || lower.includes('haul'))
      svg = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="3" width="15" height="13" rx="1"/><path d="M16 8h4l3 5v3h-7V8z"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/></svg>';
    else
      svg = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>';
    return this.sanitizer.bypassSecurityTrustHtml(svg);
  }
}

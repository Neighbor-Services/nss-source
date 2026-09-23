import { Component, OnInit, AfterViewInit, signal, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-legal',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './legal.component.html',
  styleUrl: './legal.component.css'
})
export class LegalComponent implements OnInit, AfterViewInit {
  @Input() type: 'terms' | 'privacy' = 'terms';

  constructor(private route: ActivatedRoute) {}

  ngOnInit() {
    this.route.data.subscribe(data => {
      if (data['type']) {
        this.type = data['type'];
      }
      setTimeout(() => this.setupObservers(), 100);
    });
  }

  ngAfterViewInit() {
    this.setupObservers();
  }

  scrollTo(id: string) {
    const el = document.getElementById(id);
    if (el) {
      el.scrollIntoView({ behavior: 'smooth' });
    }
  }

  private setupObservers() {
    if (typeof window === 'undefined') return;

    // Scroll reveals
    const io = new IntersectionObserver((entries) => {
      entries.forEach((e) => {
        if (e.isIntersecting) {
          e.target.classList.add('in');
          io.unobserve(e.target);
        }
      });
    }, { threshold: 0.1 });

    document.querySelectorAll('.rv').forEach((el) => io.observe(el));

    // Scroll spy for TOC
    const tocLinks = Array.from(document.querySelectorAll('.directory a'));
    const sections = Array.from(document.querySelectorAll('.sec'));
    if (sections.length && tocLinks.length) {
      let current: HTMLElement | null = null;
      const spy = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            if (current) current.classList.remove('active');
            const targetId = entry.target.id;
            const link = tocLinks.find(a => (a as HTMLElement).getAttribute('href')?.includes(targetId) || a.textContent?.toLowerCase().includes(targetId));
            if (link) {
              link.classList.add('active');
              current = link as HTMLElement;
            }
          }
        });
      }, { rootMargin: '-20% 0px -60% 0px', threshold: 0 });

      sections.forEach((s) => spy.observe(s));
    }
  }
}

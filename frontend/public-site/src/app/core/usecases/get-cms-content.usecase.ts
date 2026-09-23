import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { CMSContent } from '../domain/entities/cms.model';
import { CmsRepository } from '../repositories/cms.repository';

@Injectable({
  providedIn: 'root'
})
export class GetCMSContentUseCase {
  constructor(private cmsRepo: CmsRepository) {}

  execute(): Observable<CMSContent> {
    return this.cmsRepo.getCMSContent();
  }
}

import { Observable } from 'rxjs';
import { CMSContent } from '../domain/entities/cms.model';

export abstract class CmsRepository {
  abstract getCMSContent(): Observable<CMSContent>;
}

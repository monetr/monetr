import type { EmailTemplate } from '../templates';
import { templates } from '../templates';

export interface TemplateEntry {
  name: string;
  component: EmailTemplate;
  previewProps: Record<string, any>;
}

export const templateList: TemplateEntry[] = Object.entries(templates).map(([name, component]) => ({
  name,
  component,
  previewProps: component.PreviewProps,
}));

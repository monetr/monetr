import type { EmailTemplate } from '@monetr/emails/templates';
import { templates } from '@monetr/emails/templates';

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

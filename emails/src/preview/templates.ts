import { templates } from '../templates';
import type { EmailTemplate } from '../templates';

export interface TemplateEntry {
  name: string;
  component: EmailTemplate;
  previewProps: Record<string, any>;
}

// Derive the preview list from the shared template registry.
export const templateList: TemplateEntry[] = Object.entries(templates).map(
  ([name, component]) => ({
    name,
    component,
    previewProps: component.PreviewProps,
  }),
);

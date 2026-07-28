import { t } from 'intlayer';

const projectsContent = {
  key: 'projects',
  content: {
    title: t({
      en: 'Projects',
      uk: 'Проєкти',
    }),
    subtitle: t({
      en: 'Manage and track all your active projects',
      uk: 'Керуйте та відстежуйте усі ваші активні проєкти',
    }),
    createProjectButton: t({
      en: 'Create Project',
      uk: 'Створити проєкт',
    }),
    noProjectsFound: t({
      en: 'No projects found',
      uk: 'Проєктів не знайдено',
    }),
    noProjectsDesc: t({
      en: 'Get started by creating your first project',
      uk: 'Розпочніть, створивши свій перший проєкт',
    }),
    createNewProjectTitle: t({
      en: 'Create New Project',
      uk: 'Створити новий проєкт',
    }),
    projectNameLabel: t({
      en: 'Project Name',
      uk: 'Назва проєкту',
    }),
    projectNamePlaceholder: t({
      en: 'e.g. Apartment Renovation',
      uk: 'напр. Ремонт квартири',
    }),
    descriptionLabel: t({
      en: 'Description',
      uk: 'Опис',
    }),
    descriptionPlaceholder: t({
      en: 'Project goals, scope, and notes...',
      uk: 'Цілі проєкту, обсяг та примітки...',
    }),
    cancelButton: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
    nameRequiredError: t({
      en: 'Project name is required',
      uk: 'Назва проєкту є обов’язковою',
    }),
    projectCreatedSuccess: t({
      en: 'Project created successfully',
      uk: 'Проєкт успішно створено',
    }),
    failedToCreateProject: t({
      en: 'Failed to create project',
      uk: 'Не вдалося створити проєкт',
    }),
  },
};

export default projectsContent;

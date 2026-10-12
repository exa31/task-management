export const settings = {
  title: 'Account & Module Settings',
  subtitle: 'Manage your profile, OAuth bindings, project task modules, and account preferences.',

  // Profile Section
  profile: {
    title: 'User Profile',
    changePhoto: 'Change',
    uploadCropWebp: 'Upload & Auto Crop WebP',
    fullName: 'Full Name',
    email: 'Email Address (Primary Account)',
    saveChanges: 'Save Changes',
    saving: 'Saving...',
    updateSuccess: 'Profile updated successfully.',
    updateError: 'Failed to update profile.'
  },

  // OAuth Bindings Section
  oauth: {
    title: 'OAuth Account Connections',
    subtitle: 'Link your Google or Apple account for 1-click sign-in across devices.',
    google: 'Google',
    apple: 'Apple ID',
    connected: 'Connected',
    notLinked: 'Not linked',
    disconnect: 'Disconnect',
    linkGoogle: 'Link Google',
    linkApple: 'Link Apple',
    googleSdkNotReady: 'Google SDK is not ready. Please check your internet connection.',
    googleClientNotConfigured: 'Google Client ID is not configured.',
    googleAuthCancelled: 'Google authorization was cancelled or failed.',
    googleLinkedSuccess: 'Google account linked successfully.',
    googleLinkError: 'Failed to link Google account.',
    googleInitError: 'Failed to open Google sign-in.',
    appleSdkNotReady: 'Apple Sign-In SDK is not ready. Please check your internet connection.',
    appleLinkedSuccess: 'Apple account linked successfully.',
    appleLinkError: 'Failed to link Apple account.',
    unbindSuccess: '{provider} account unlinked successfully.',
    unbindError: 'Failed to unlink account.'
  },

  // Project Task Modules Section
  projects: {
    title: 'Task Modules / Projects',
    subtitle: 'Group tasks into specific modules/projects (e.g. ftracker, e-commerce).',
    addProject: 'Add Project',
    loading: 'Loading projects...',
    empty: 'No project modules yet. Create a new project to organize your tasks!',
    tasksConnected: '{count} tasks connected',
    editModalTitle: 'Edit Project Module',
    createModalTitle: 'Add Project Module',
    nameLabel: 'Project Name',
    namePlaceholder: 'e.g. ftracker, e-commerce',
    descLabel: 'Description (Optional)',
    descPlaceholder: 'Short project description...',
    colorLabel: 'Theme Color',
    cancel: 'Cancel',
    save: 'Save',
    create: 'Create Project',
    deleteHeader: 'Delete Project Module',
    deleteMessage: 'Delete project "{name}"? Linked tasks will not be deleted.',
    deleteAccept: 'Delete',
    deleteReject: 'Cancel'
  },

  // MCP AI Link Section
  mcp: {
    title: 'Model Context Protocol (MCP) & AI',
    subtitle:
      'Generate Personal Access Tokens to let AI coding agents (Antigravity, Claude, Cline) interact directly with your workspace tasks and notes.',
    button: 'Open MCP Settings'
  },

  // Notification & Due Date Warning Preferences
  notifications: {
    title: 'Notification & Due Date Warning Preferences',
    subtitle:
      'Configure your timezone, deadline warning threshold alerts, and automated email reminders.',
    timezoneLabel: 'Display & Reminder Timezone',
    timezoneHint: 'All due date alerts and email summaries will be aligned with this timezone.',
    thresholdLabel: 'Due Date Warning Alert Threshold',
    thresholdHint:
      'TaskHub will send an email alert (Resend), realtime notification (SSE), and mobile push (FCM) when a task approaches its deadline.',
    threshold15m: '15 Minutes before deadline',
    threshold1h: '1 Hour before deadline',
    threshold3h: '3 Hours before deadline',
    threshold12h: '12 Hours before deadline',
    threshold24h: '24 Hours (1 Day) before deadline',
    threshold48h: '48 Hours (2 Days) before deadline',
    emailNotificationsLabel: 'Email Deadline Warnings (via Resend)',
    emailNotificationsDesc:
      'Send an email notification to your account when tasks reach the warning threshold.',
    savePreferences: 'Save Notification Preferences',
    saving: 'Saving preferences...',
    savedSuccess: 'Notification preferences saved successfully.',
    savedError: 'Failed to update notification preferences.'
  },

  // Danger Zone Section
  dangerZone: {
    title: 'Danger Zone',
    subtitle:
      'This action cannot be undone. Deleting your account permanently clears all data, including tasks, subtasks, notes, avatar files in storage, and active sessions.',
    deleteAccountBtn: 'Delete Account Permanently',
    modalTitle: 'Permanently Delete Account?',
    modalDesc:
      'This action will permanently delete all account data from the database, including all tasks, projects, notes, and avatar images in storage, and revoke all active sessions. This cannot be undone.',
    typeDeletePrompt: 'Type DELETE to confirm:',
    errorConfirmation: 'Please type DELETE to confirm.',
    cancelBtn: 'Cancel',
    confirmDeleteBtn: 'I understand, permanently delete',
    deleting: 'Deleting account permanently...',
    deletedToast: 'Your account has been permanently deleted.',
    deleteErrorToast: 'Failed to delete account.'
  },

  // Avatar Crop Modal
  avatarModal: {
    title: 'Profile Photo',
    subtitle: '1:1 Auto-crop & automatic WebP conversion',
    uploadPrompt: 'Choose or drag a photo here',
    supportedFormats: 'PNG, JPG, JPEG, WEBP (max. 10MB)',
    changePhoto: 'Change Photo',
    resetPosition: 'Reset Position',
    cancel: 'Cancel',
    save: 'Save Avatar',
    saving: 'Compressing & saving...',
    invalidFileType: 'Please choose an image file (JPG, PNG, or WEBP).',
    fileTooLarge: 'Maximum file size is 10MB.',
    convertError: 'Failed to convert image to WebP format.',
    uploadError: 'Failed to upload avatar.',
    uploadSuccess: 'Profile photo updated successfully.'
  }
}

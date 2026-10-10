export const notes = {
  title: 'Idea Scratchpad ({count})',
  subtitle:
    'Capture fleeting thoughts, snippets, or plans. Convert any idea to a task with one click.',
  allIdeas: 'All Ideas ({count})',
  writeNewIdea: 'Write New Idea',
  loadingIdeas: 'Loading ideas...',
  failedLoad: 'Failed to Load Ideas',
  noIdeasTitle: 'No ideas saved yet',
  noIdeasSubtitle:
    'Use this scratchpad to quickly log sudden thoughts. Convert any note into a full task with 1-click!',
  captureFirst: 'Capture First Idea',
  convertToTask: 'Convert to Task',
  converted: 'Task #{id}',
  pin: 'Pin to top',
  unpin: 'Unpin note',
  deleteConfirmHeader: 'Delete Idea Note',
  deleteConfirmMessage: 'Are you sure you want to delete this note? This action cannot be undone.',
  createdSuccess: 'Note created successfully',
  updatedSuccess: 'Note updated successfully',
  deletedSuccess: 'Note deleted successfully',
  convertedSuccess: 'Note converted to task successfully',
  typeKnowledge: 'Knowledge',
  typeActionable: 'Actionable',
  tagsLabel: 'Tags',
  allTags: 'All Tags',
  howToStorage: 'How-To / Storage',
  statsTotal: 'Total Notes',
  statsActionable: 'Actionable',
  statsKnowledge: 'Knowledge Base',
  statsPinned: 'Pinned'
}

export const noteModal = {
  createTitle: 'Write New Idea Note',
  editTitle: 'Edit Idea Note',
  subtitle: 'Jot down raw thoughts, drafts, or project brain-dumps.',
  titleLabel: 'Note Title',
  titlePlaceholder: 'e.g. Real-time websocket notification architecture',
  titleRequired: 'Note title is required',
  contentLabel: 'Content / Thoughts',
  contentPlaceholder: 'Outline thoughts, reference links, code snippets...',
  colorTheme: 'Card Color',
  tagsLabel: 'Tags',
  tagsPlaceholder: 'research, v2, ux (comma separated)',
  typeLabel: 'Note Type',
  actionableTitle: 'Actionable Idea',
  actionableDesc: 'Can be converted into a task / backlog',
  knowledgeTitle: 'Knowledge / How-To',
  knowledgeDesc: 'Reference & guide (storage only)',
  pinLabel: 'Pin this note to the top',
  save: 'Save Note',
  colors: {
    default: 'Slate',
    indigo: 'Indigo',
    purple: 'Purple',
    blue: 'Blue',
    green: 'Emerald',
    amber: 'Amber',
    rose: 'Rose'
  }
}

export const convertModal = {
  title: 'Convert Idea to Task',
  subtitle: 'Upgrade this scratchpad note into an actionable task pipeline item.',
  fromIdea: 'From Idea Note:',
  priorityLabel: 'Priority Level',
  pinLabel: "Pin directly to Today's Focus Horizon",
  checklistPlaceholder:
    'Example:\nStep 1: Research documentation\nStep 2: Build prototype\nStep 3: Review and ship',
  cancel: 'Cancel',
  submit: 'Convert & Create Task'
}

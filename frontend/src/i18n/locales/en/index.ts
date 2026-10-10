import { login } from './auth'
import { commandPalette } from './command'
import { common } from './common'
import { dashboard, focus } from './dashboard'
import { errors } from './errors'
import { landing } from './landing'
import { mcp } from './mcp'
import { nav } from './nav'
import { convertModal, noteModal, notes } from './notes'
import { projects } from './projects'
import { reminders } from './reminders'
import { settings } from './settings'
import { taskDrawer, taskModal, tasks } from './tasks'

export const en = {
  common,
  nav,
  landing,
  dashboard,
  focus,
  tasks,
  taskModal,
  taskDrawer,
  projects,
  reminders,
  notes,
  noteModal,
  convertModal,
  mcp,
  commandPalette,
  login,
  settings,
  errors
}

export type TranslationSchema = typeof en

export default en

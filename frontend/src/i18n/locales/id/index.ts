import type { TranslationSchema } from '../en/index'
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
import { calendar } from './calendar'
import { settings } from './settings'
import { taskDrawer, taskModal, tasks } from './tasks'

export const id: TranslationSchema = {
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
  calendar,
  notes,
  noteModal,
  convertModal,
  mcp,
  commandPalette,
  login,
  settings,
  errors
}

export default id

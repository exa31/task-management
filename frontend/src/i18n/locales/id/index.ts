import type { TranslationSchema } from '../en/index'
import { login } from './auth'
import { commandPalette } from './command'
import { common } from './common'
import { dashboard, focus } from './dashboard'
import { errors } from './errors'
import { mcp } from './mcp'
import { nav } from './nav'
import { convertModal, noteModal, notes } from './notes'
import { taskDrawer, taskModal, tasks } from './tasks'

export const id: TranslationSchema = {
  common,
  nav,
  dashboard,
  focus,
  tasks,
  taskModal,
  taskDrawer,
  notes,
  noteModal,
  convertModal,
  mcp,
  commandPalette,
  login,
  errors,
}

export default id

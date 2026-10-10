import { Column } from 'primereact/column'
import { DataTable } from 'primereact/datatable'
import { useState } from 'react'

export default function Users() {
  const [users] = useState([
    { id: 1, name: 'Alice Smith', email: 'alice@example.com', role: 'Admin', status: 'Active' },
    { id: 2, name: 'Bob Jones', email: 'bob@example.com', role: 'User', status: 'Inactive' }
  ])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Users</h1>
          <p className="text-slate-500 text-sm mt-1">Manage team members and their roles.</p>
        </div>
      </div>

      <div className="bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm">
        <DataTable value={users} responsiveLayout="scroll">
          <Column field="name" header="Name" />
          <Column field="email" header="Email" />
          <Column field="role" header="Role" />
          <Column field="status" header="Status" />
        </DataTable>
      </div>
    </div>
  )
}

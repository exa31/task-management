import { Button } from 'primereact/button'
import { Link } from 'react-router-dom'

export default function NotFound() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-50 text-center">
      <div className="max-w-md w-full">
        <div className="text-indigo-500 font-bold text-9xl tracking-tighter mb-4 opacity-50 dark:opacity-20">
          404
        </div>
        <h1 className="text-3xl font-bold tracking-tight mb-3">Page not found</h1>
        <p className="text-slate-500 dark:text-slate-400 mb-8">
          Sorry, we couldn't find the page you're looking for. It might have been moved or deleted.
        </p>
        <Link to="/">
          <Button label="Back to Dashboard" icon="pi pi-home" className="bg-indigo-600 hover:bg-indigo-700 border-none px-6 py-2.5 rounded-xl font-medium" />
        </Link>
      </div>
    </div>
  )
}

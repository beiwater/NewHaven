import { useState, type FormEvent } from 'react'
import { useSubmitReport, type ReportCategory, CATEGORY_LABELS } from '@/api/report.api'

const CATEGORIES: ReportCategory[] = ['bug', 'feature', 'feedback', 'other']

export function ReportPanel() {
  const [open, setOpen] = useState(false)
  const [category, setCategory] = useState<ReportCategory>('bug')
  const [description, setDescription] = useState('')
  const submit = useSubmitReport()

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!description.trim()) return
    submit.mutate(
      { category, description: description.trim() },
      {
        onSuccess: () => {
          setDescription('')
          setOpen(false)
        },
      },
    )
  }

  const handleClose = () => {
    if (submit.isPending) return
    setOpen(false)
  }

  return (
    <>
      {/* Floating trigger button — above the chat button */}
      {!open && (
        <button
          onClick={() => setOpen(true)}
          className="fixed bottom-[170px] right-4 z-50 w-10 h-10 bg-red-700 hover:bg-red-800 text-white rounded-full shadow-lg flex items-center justify-center transition-colors"
          title="Report a problem"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </button>
      )}

      {/* Report dialog */}
      {open && (
        <div className="fixed bottom-[170px] right-4 z-50 w-80 bg-white border-2 border-red-200 rounded-lg shadow-xl flex flex-col overflow-hidden">
          {/* Header */}
          <div className="flex items-center justify-between px-3 py-2 bg-red-700 text-white shrink-0">
            <span className="text-xs font-semibold">Report a Problem</span>
            <button
              onClick={handleClose}
              disabled={submit.isPending}
              className="text-red-200 hover:text-white disabled:opacity-50"
            >
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Content */}
          <form onSubmit={handleSubmit} className="p-3 space-y-3">
            {/* Category selector */}
            <div>
              <label className="text-[10px] uppercase tracking-wider text-red-600 font-semibold block mb-1.5">
                Category
              </label>
              <div className="flex flex-wrap gap-1.5">
                {CATEGORIES.map((cat) => (
                  <button
                    key={cat}
                    type="button"
                    onClick={() => setCategory(cat)}
                    className={`px-2.5 py-1 rounded text-[10px] font-bold transition-colors ${
                      category === cat
                        ? 'bg-red-700 text-white'
                        : 'bg-red-100 text-red-700 hover:bg-red-200'
                    }`}
                  >
                    {CATEGORY_LABELS[cat]}
                  </button>
                ))}
              </div>
            </div>

            {/* Description */}
            <div>
              <label className="text-[10px] uppercase tracking-wider text-red-600 font-semibold block mb-1.5">
                Description
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value.slice(0, 2000))}
                placeholder="Describe the issue or suggestion..."
                disabled={submit.isPending}
                rows={4}
                className="w-full px-2 py-1.5 text-xs bg-white border border-red-300 rounded text-red-900 placeholder-red-400 disabled:opacity-50 resize-none"
              />
              <div className="flex justify-end mt-0.5">
                <span className={`text-[9px] font-medium ${
                  description.length > 1800 ? 'text-red-500' : description.length > 1500 ? 'text-amber-500' : 'text-red-400'
                }`}>
                  {description.length} / 2000
                </span>
              </div>
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={submit.isPending || !description.trim()}
              className="w-full py-1.5 bg-red-700 hover:bg-red-800 disabled:bg-red-400 text-white text-xs font-semibold rounded transition-colors disabled:cursor-not-allowed"
            >
              {submit.isPending ? 'Submitting...' : submit.isSuccess ? 'Submitted ✓' : 'Submit Report'}
            </button>
            {submit.isError && (
              <p className="text-[10px] text-red-500 text-center">{submit.error?.message || 'Failed to submit. Try again.'}</p>
            )}
            {submit.isSuccess && (
              <p className="text-[10px] text-green-600 text-center">Report submitted. Thank you!</p>
            )}

            {/* GitHub issues link */}
            <div className="pt-2 border-t border-red-100">
              <a
                href="https://github.com/beiwater/NewHaven/issues"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center justify-center gap-1.5 text-[10px] text-red-500 hover:text-red-700 transition-colors"
              >
                <svg className="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z" />
                </svg>
                <span>Report on GitHub Issues</span>
              </a>
            </div>
          </form>
        </div>
      )}
    </>
  )
}

import type { Overview } from '@/service/api/overview';

export interface OverviewLoadState {
  data: Overview | null;
  loading: boolean;
  failed: boolean;
}

/** Share one load path for entry, cached-page activation and manual refresh. */
export function createOverviewLoader(state: OverviewLoadState, fetch: () => Promise<Overview | null>) {
  let sequence = 0;

  async function reload() {
    // Mount and initial activation happen together for cached pages.
    if (state.loading) return;
    const current = ++sequence;
    state.loading = true;
    state.failed = false;
    try {
      const data = await fetch();
      if (current !== sequence) return;
      if (data) state.data = data;
      state.failed = data === null;
    } catch {
      if (current === sequence) state.failed = true;
    } finally {
      if (current === sequence) state.loading = false;
    }
  }

  function invalidate() {
    // Ignore a response that finishes after leaving this page.
    sequence += 1;
    state.loading = false;
  }

  return { reload, invalidate };
}

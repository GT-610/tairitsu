// @ts-nocheck -- happy-dom installs its own global DOM types, which cannot be
// unified with the DOM lib types this project compiles against.
import { GlobalRegistrator } from '@happy-dom/global-registrator'
import { cleanup } from '@testing-library/react'
import { afterAll, afterEach, beforeAll } from 'bun:test'
import { act } from 'react'

/**
 * Installs a DOM for tests that need to render and interact with a component.
 *
 * A static server render only exercises whichever branch happens to be active
 * on first paint, so anything that depends on state or an event cannot be
 * asserted that way. Call this at the top level of a test file and query through
 * the object `render` returns rather than the exported `screen`, which binds to
 * `document.body` when it is first imported and therefore before the DOM below
 * exists. `user-event` has the same problem and needs `{ document }` passed to
 * `setup`.
 *
 * Call it from the module body, not inside a `describe`, so the hooks register
 * for the whole file.
 */
export function useTestDOM() {
  beforeAll(() => {
    GlobalRegistrator.register({ url: 'http://localhost/' })
    globalThis.IS_REACT_ACT_ENVIRONMENT = true
  })

  afterEach(async () => {
    // Unmount through act so React finishes the teardown while the DOM is still
    // installed; a scheduler callback that lands after the window is gone
    // throws instead.
    await act(async () => {
      cleanup()
    })
  })

  afterAll(async () => {
    // React schedules follow-up work through a task queue. Leaving act mode
    // first makes it flush that work synchronously rather than queuing it into
    // a window that is about to disappear.
    globalThis.IS_REACT_ACT_ENVIRONMENT = false
    await GlobalRegistrator.unregister()
  })
}

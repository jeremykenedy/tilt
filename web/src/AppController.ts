import HUD from "./HUD"

// A Websocket that automatically retries.

class AppController {
  url: string
  loadCount: number
  liveSocket: boolean
  tryConnectCount: number
  // TOOD(dmiller): optional type?
  socket: WebSocket | null = null
  component: HUD
  disposed: boolean = false

  /**
   * @param url The url of the websocket to pull data from
   * @param component The top-level component for the app.
   *     Has one method, setAppState, that sets the global state of the
   *     app. This state has two properties
   *     - Message (string): A status message about the state of the socket
   *     - View (Object): A JSON serialization of the Go struct in internal/renderer/view
   */
  constructor(url: string, component: HUD) {
    if (!component.setAppState) {
      throw new Error("App component has no setAppState method")
    }

    this.url = url
    this.component = component
    this.tryConnectCount = 0
    this.liveSocket = false
    this.loadCount = 0
  }

  createNewSocket() {
    this.tryConnectCount++
    this.socket = new WebSocket(this.url)
    this.socket.addEventListener("close", this.onSocketClose.bind(this))
    this.socket.addEventListener("message", event => {
      if (!this.liveSocket) {
        this.loadCount++
      }
      this.liveSocket = true
      this.tryConnectCount = 0

      let data = JSON.parse(event.data)
      // @ts-ignore
      this.component.setAppState({ View: data })
    })
  }

  dispose() {
    this.disposed = true
    if (this.socket) {
      this.socket.close()
    }
  }

  onSocketClose() {
    let wasAlive = this.liveSocket
    this.liveSocket = false
    if (this.disposed) {
      return
    }

    if (wasAlive) {
      this.component.setAppState({
        View: null,
        Message: "Disconnected…",
        IsSidebarClosed: false,
      })
      this.createNewSocket()
      return
    }

    let backoff = Math.pow(2, this.tryConnectCount) * 1000
    let maxTimeout = 10 * 1000 // 10sec
    let isLocal = this.url.indexOf("ws://localhost") == 0
    if (isLocal) {
      // if this is a local connection, max out at 1.5sec.
      // this makes it a bit easier to detect when a window is already open.
      maxTimeout = 1500
    }
    let timeout = Math.min(maxTimeout, backoff)

    setTimeout(() => {
      if (this.disposed) {
        return
      }
      let message = this.loadCount ? "Reconnecting…" : "Loading…"
      this.component.setAppState({
        View: null,
        Message: message,
        IsSidebarClosed: false,
      })
      this.createNewSocket()
    }, timeout)
  }
}

export default AppController

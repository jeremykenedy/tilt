import React, { PureComponent } from "react"
import { ReactComponent as ChevronSvg } from "./assets/svg/chevron.svg"
import { Link } from "react-router-dom"
import { combinedStatus, warnings } from "./status"
import "./Sidebar.scss"
import { ResourceView, TriggerMode, RuntimeStatus, Build } from "./types"
import TimeAgo from "react-timeago"
import { isZeroTime } from "./time"
import PathBuilder from "./PathBuilder"
import { timeAgoFormatter } from "./timeFormatters"
import { AlertResource } from "./AlertPane"
import SidebarIcon from "./SidebarIcon"
import SidebarTriggerButton from "./SidebarTriggerButton"

class SidebarItem {
  name: string
  status: RuntimeStatus
  hasWarnings: boolean
  hasEndpoints: boolean
  lastDeployTime: string
  pendingBuildSince: string
  currentBuildStartTime: string
  alertResource: AlertResource
  triggerMode: TriggerMode
  hasPendingChanges: boolean
  lastBuild: Build | null = null

  /**
   * Create a pared down SidebarItem from a ResourceView
   */
  constructor(res: any) {
    this.name = res.Name
    this.status = combinedStatus(res)
    this.hasWarnings = warnings(res).length > 0
    this.hasEndpoints = (res.Endpoints || []).length
    this.lastDeployTime = res.LastDeployTime
    this.pendingBuildSince = res.PendingBuildSince
    this.currentBuildStartTime = res.CurrentBuild.StartTime
    this.alertResource = new AlertResource(res)
    this.triggerMode = res.TriggerMode
    this.hasPendingChanges = res.HasPendingChanges
    let buildHistory = res.BuildHistory || []
    if (buildHistory.length > 0) {
      this.lastBuild = buildHistory[0]
    }
  }

  numberOfAlerts(): number {
    return this.alertResource.numberOfAlerts()
  }
}

type SidebarProps = {
  isClosed: boolean
  items: SidebarItem[]
  selected: string
  toggleSidebar: any
  resourceView: ResourceView
  pathBuilder: PathBuilder
}

class Sidebar extends PureComponent<SidebarProps> {
  render() {
    let pb = this.props.pathBuilder
    let classes = ["Sidebar"]
    if (this.props.isClosed) {
      classes.push("is-closed")
    }

    let allItemClasses = "resLink resLink--all"
    if (!this.props.selected) {
      allItemClasses += " is-selected"
    }
    let allLink =
      this.props.resourceView === ResourceView.Alerts
        ? pb.path("/alerts")
        : pb.path("/")
    let totalAlerts = this.props.items
      .map(i => i.numberOfAlerts())
      .reduce((sum, current) => sum + current, 0)

    let allItem = (
      <li>
        <Link className={allItemClasses} to={allLink}>
          <span className="resLink--all-name">All</span>
          {totalAlerts > 0 ? (
            <span className="resLink-alertBadge">{totalAlerts}</span>
          ) : (
            ""
          )}
          <span className="resLink-timeAgo empty">—</span>
          <span className="resLink-isDirty" />
        </Link>
      </li>
    )

    let listItems = this.props.items.map(item => {
      let link = `/r/${item.name}`
      if (this.props.resourceView === ResourceView.Preview) {
        link += "/preview"
      } else if (this.props.resourceView === ResourceView.Alerts) {
        link += "/alerts"
      }

      let formatter = timeAgoFormatter
      let hasBuilt = !isZeroTime(item.lastDeployTime)
      let building = !isZeroTime(item.currentBuildStartTime)
      let timeAgo = <TimeAgo date={item.lastDeployTime} formatter={formatter} />
      let isSelected = this.props.selected === item.name
      let isManualTriggerMode =
        item.triggerMode === TriggerMode.TriggerModeManual

      let classes = "resLink"
      if (building) {
        classes += " resLink--building"
      }

      if (isSelected) {
        classes += " is-selected"
      }
      return (
        <li key={item.name}>
          <SidebarTriggerButton
            isSelected={isSelected}
            resourceName={item.name}
            isReady={item.hasPendingChanges && !building}
            triggerMode={item.triggerMode}
          />
          <Link className={classes} to={pb.path(link)}>
            <div className="sidebarIcon">
              <SidebarIcon
                status={item.status}
                triggerMode={item.triggerMode}
                hasWarning={item.hasWarnings}
                isBuilding={building}
                isDirty={item.hasPendingChanges}
                lastBuild={item.lastBuild}
              />
            </div>
            <p className="resLink-name" title={item.name}>
              {item.name}
            </p>
            {item.numberOfAlerts() > 0 ? (
              <span className="resLink-alertBadge">
                {item.numberOfAlerts()}
              </span>
            ) : (
              ""
            )}
            <span className={`resLink-timeAgo ${hasBuilt ? "" : "empty"}`}>
              {hasBuilt ? timeAgo : "—"}
            </span>
            <span className="resLink-isDirty">
              {item.hasPendingChanges && isManualTriggerMode ? "*" : null}
            </span>
          </Link>
        </li>
      )
    })

    return (
      <section className={classes.join(" ")}>
        <nav className="Sidebar-resources">
          <ul className="Sidebar-list">
            {allItem}
            {listItems}
          </ul>
        </nav>
        <div className="Sidebar-spacer">&nbsp;</div>
        <button className="Sidebar-toggle" onClick={this.props.toggleSidebar}>
          <ChevronSvg /> Collapse
        </button>
      </section>
    )
  }
}

export default Sidebar

export { SidebarItem }

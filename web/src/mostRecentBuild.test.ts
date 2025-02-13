import mostRecentBuildToDisplay, { ResourceWithBuilds } from "./mostRecentBuild"
import { zeroTime } from "./time"

it("returns null if there are no builds", () => {
  const resources: Array<ResourceWithBuilds> = []

  let actual = mostRecentBuildToDisplay(resources)
  expect(actual).toBeNull()
})

it("returns the most recent build if there are no pending builds", () => {
  let recent = {
    Edits: ["main.go"],
    Error: null,
    StartTime: "2019-04-24T13:08:41.017623-04:00",
    FinishTime: "2019-04-24T13:08:42.926608-04:00",
    Log: "",
    IsCrashRebuild: false,
    Warnings: [],
  }
  let expectedTuple = {
    name: "snack",
    edits: ["main.go"],
    since: recent.StartTime,
  }
  const resource: ResourceWithBuilds = {
    name: "snack",
    buildHistory: [
      {
        Edits: ["main.go"],
        Error: null,
        StartTime: "2019-04-24T13:08:39.017623-04:00",
        FinishTime: "2019-04-24T13:08:40.926608-04:00",
        Log: "",
        IsCrashRebuild: false,
        Warnings: [],
      },
      recent,
    ],
    pendingBuildEdits: null,
    pendingBuildSince: zeroTime,
  }
  const resources: Array<ResourceWithBuilds> = [resource]

  let actual = mostRecentBuildToDisplay(resources)
  expect(actual).toEqual(expectedTuple)
})

it("returns null if there are no pending builds and the most recent build has no edits", () => {
  let recent = {
    Edits: null,
    Error: null,
    StartTime: "2019-04-24T13:08:41.017623-04:00",
    FinishTime: "2019-04-24T13:08:42.926608-04:00",
    Log: "",
    IsCrashRebuild: false,
    Warnings: [],
  }
  let expectedTuple = {
    name: "snack",
    edits: ["main.go"],
    since: recent.StartTime,
  }
  const resource: ResourceWithBuilds = {
    name: "snack",
    buildHistory: [
      {
        Edits: null,
        Error: null,
        StartTime: "2019-04-24T13:08:39.017623-04:00",
        FinishTime: "2019-04-24T13:08:40.926608-04:00",
        Log: "",
        IsCrashRebuild: false,
        Warnings: [],
      },
      recent,
    ],
    pendingBuildEdits: null,
    pendingBuildSince: zeroTime,
  }
  const resources: Array<ResourceWithBuilds> = [resource]

  let actual = mostRecentBuildToDisplay(resources)
  expect(actual).toBeNull()
})

it("returns the pending build if there is one", () => {
  let expectedTuple = {
    name: "snack",
    edits: ["bar"],
    since: "2019-04-24T13:08:41.017623-04:00",
  }
  const resource: ResourceWithBuilds = {
    name: "snack",
    buildHistory: [
      {
        Edits: null,
        Error: null,
        StartTime: "2019-04-24T13:08:39.017623-04:00",
        FinishTime: "2019-04-24T13:08:40.926608-04:00",
        Log: "",
        IsCrashRebuild: false,
        Warnings: [],
      },
    ],
    pendingBuildEdits: ["bar"],
    pendingBuildSince: "2019-04-24T13:08:41.017623-04:00",
  }
  const resources = [resource]

  let actual = mostRecentBuildToDisplay(resources)
  expect(actual).toEqual(expectedTuple)
})

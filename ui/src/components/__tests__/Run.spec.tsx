import * as React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import { Run, Props } from "../Run"
import { Run as RunType, RunStatus } from "../../types"
import { RequestStatus } from "../Request"

export type RunInstance = {}

const MockRun: RunType = {
  instance: {
    dns_name: "dns_name",
    instance_id: "instance_id",
  },
  task_arn: "task_arn",
  run_id: "run_id",
  definition_id: "definition_id",
  alias: "alias",
  image: "image",
  cluster: "cluster",
  exit_code: 0,
  status: RunStatus.RUNNING,
  started_at: "started_at",
  finished_at: "finished_at",
  group_name: "group_name",
  env: [],
}

const Proxy: React.FunctionComponent<Props> = props => (
  <MemoryRouter>
    <Run {...props} />
  </MemoryRouter>
)

describe("Run", () => {
  const realSet = Run.prototype.setRequestInterval
  const realClear = Run.prototype.clearRequestInterval

  beforeEach(() => {
    Run.prototype.setRequestInterval = jest.fn()
    Run.prototype.clearRequestInterval = jest.fn()
  })

  afterEach(() => {
    Run.prototype.setRequestInterval = realSet
    Run.prototype.clearRequestInterval = realClear
  })

  /**
   * If the run is in a non-stopped state, the component should start an
   * interval to continuously fetch the run.
   */
  it("sets a request interval if the run isn't stopped on componentDidMount", () => {
    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(0)

    // Mount a stopped run.
    mount(
      <MemoryRouter>
        <Run
          requestStatus={RequestStatus.READY}
          data={{
            ...MockRun,
            status: RunStatus.STOPPED,
          }}
          isLoading={false}
          error={null}
          runID="a"
          request={jest.fn()}
        />
      </MemoryRouter>
    )
    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(0)

    // Mount a running one.
    mount(
      <MemoryRouter>
        <Run
          requestStatus={RequestStatus.READY}
          data={MockRun}
          isLoading={false}
          error={null}
          runID="a"
          request={jest.fn()}
        />
      </MemoryRouter>
    )
    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(1)
  })

  it("sets the request interval if props.requestStatus changes from NOT_READY to READY and the run is not stopped.", () => {
    // Request has not completed.
    const stoppedWrapper = mount(
      <Proxy
        requestStatus={RequestStatus.NOT_READY}
        data={null}
        isLoading={false}
        error={null}
        runID="a"
        request={jest.fn()}
      />
    )
    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(0)

    // Set requestStatus to READY.
    stoppedWrapper.setProps({
      requestStatus: RequestStatus.READY,
      data: {
        ...MockRun,
        status: RunStatus.STOPPED,
      },
    })

    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(0)

    // Request has not completed.
    const runningWrapper = mount<Run>(
      <Proxy
        requestStatus={RequestStatus.NOT_READY}
        data={null}
        isLoading={false}
        error={null}
        runID="a"
        request={jest.fn()}
      />
    )
    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(0)

    // Set requestStatus to READY.
    runningWrapper.setProps({
      requestStatus: RequestStatus.READY,
      data: {
        ...MockRun,
        status: RunStatus.RUNNING,
      },
    })

    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(1)
  })

  it("clears the request interval if the run transitions into a stopped state on componentDidUpdate", () => {
    const wrapper = mount(
      <Proxy
        requestStatus={RequestStatus.READY}
        data={MockRun}
        isLoading={false}
        error={null}
        runID="a"
        request={jest.fn()}
      />
    )
    expect(Run.prototype.clearRequestInterval).toHaveBeenCalledTimes(0)
    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(1)

    // Set the state to stopped
    wrapper.setProps({
      data: {
        ...MockRun,
        status: RunStatus.STOPPED,
      },
    })

    expect(Run.prototype.clearRequestInterval).toHaveBeenCalledTimes(1)
  })

  it("clears the request interval on componentWillUnmount", async () => {
    const wrapper = mount(
      <MemoryRouter>
        <Run
          requestStatus={RequestStatus.READY}
          data={MockRun}
          isLoading={false}
          error={null}
          runID="a"
          request={jest.fn()}
        />
      </MemoryRouter>
    )
    expect(Run.prototype.clearRequestInterval).toHaveBeenCalledTimes(0)
    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(1)
    wrapper.unmount()
    expect(Run.prototype.clearRequestInterval).toHaveBeenCalledTimes(1)
  })
})

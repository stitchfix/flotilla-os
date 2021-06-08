import * as React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import { Run, Props } from "../Run"
import {
  Run as RunType,
  RunStatus,
  ExecutionEngine,
  NodeLifecycle,
} from "../../types"
import { RequestStatus } from "../Request"
import { Provider } from "react-redux"
import store from "../../state/store"

jest.mock("../../workers/index")

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
  started_at: "2019-10-24T05:21:51",
  finished_at: "2019-10-25T06:21:51",
  group_name: "group_name",
  env: [],
  engine: ExecutionEngine.K8S,
  node_lifecycle: NodeLifecycle.ON_DEMAND,
  max_cpu_used: 0,
  max_memory_used: 0,
  pod_name: "",
  cpu: 100,
  memory: 100,
  queued_at: "2019-10-24T04:21:51",
}

const Proxy: React.FunctionComponent<Props> = props => (
  <Provider store={store}>
    <MemoryRouter>
      <Run {...props} />
    </MemoryRouter>
  </Provider>
)

const defaultProps: Props = {
  requestStatus: RequestStatus.READY,
  data: MockRun,
  isLoading: false,
  error: null,
  runID: MockRun.run_id,
  request: jest.fn(),
  query: {},
  setQuery: jest.fn(),
  receivedAt: new Date(),
}

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
      <Proxy
        {...defaultProps}
        data={{
          ...MockRun,
          status: RunStatus.STOPPED,
        }}
      />
    )
    expect(Run.prototype.setRequestInterval).toHaveBeenCalledTimes(0)

    // Mount a running one.
    mount(<Proxy {...defaultProps} />)
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
        query={{}}
        setQuery={jest.fn()}
        receivedAt={new Date()}
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
        query={{}}
        setQuery={jest.fn()}
        receivedAt={new Date()}
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
        query={{}}
        setQuery={jest.fn()}
        receivedAt={new Date()}
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
})

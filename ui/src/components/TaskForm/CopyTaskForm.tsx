import * as React from "react"
import { get } from "lodash"
import { flotillaUIRequestStates } from "../../types"
import TaskContext from "../Task/TaskContext"
import CreateTaskForm from "./CreateTaskForm"
import Loader from "../styled/Loader"
import config from "../../config"

/**
 * The CopyTaskForm is just a CreateTaskForm that gets its defaults from the
 * task it's copying from (via the TaskContext).
 */
const CopyTaskForm: React.SFC<{}> = () => (
  <TaskContext.Consumer>
    {ctx => {
      if (ctx.requestState === flotillaUIRequestStates.READY) {
        return (
          <CreateTaskForm
            defaultValues={{
              alias: "",
              command: get(ctx, ["data", "command"], ""),
              env: get(ctx, ["data", "env"], []),
              group_name: get(ctx, ["data", "group_name"], ""),
              image: get(ctx, ["data", "image"], config.IMAGE_PREFIX),
              memory: get(ctx, ["data", "memory"], 1024),
              tags: get(ctx, ["data", "tags"], []),
            }}
            title={`Copy Task ${get(ctx, ["data", "alias"], ctx.definitionID)}`}
          />
        )
      }

      return <Loader />
    }}
  </TaskContext.Consumer>
)

export default CopyTaskForm

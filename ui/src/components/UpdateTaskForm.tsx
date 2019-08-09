import * as React from "react"
import { RouteComponentProps } from "react-router-dom"
import { Button, Intent, Spinner } from "@blueprintjs/core"
import { Formik, Form, FormikProps } from "formik"
import { get } from "lodash"
import * as Yup from "yup"
import api from "../api"
import { UpdateTaskPayload, Task } from "../types"
import Request, {
  ChildProps as RequestChildProps,
  RequestStatus,
} from "./Request"
import BaseTaskForm, {
  validationSchema as baseTaskFormValidationSchema,
} from "./BaseTaskForm"
import { TaskContext, TaskCtx } from "./Task"
import ErrorCallout from "./ErrorCallout"
import Toaster from "./Toaster"

export const validationSchema = Yup.object().shape(baseTaskFormValidationSchema)

export type Props = Pick<
  FormikProps<UpdateTaskPayload>,
  "values" | "setFieldValue" | "isValid" | "errors"
> &
  Pick<
    RequestChildProps<Task, { data: UpdateTaskPayload }>,
    "requestStatus" | "error" | "isLoading"
  >

export const UpdateTaskForm: React.FunctionComponent<Props> = ({
  values,
  isValid,
  setFieldValue,
  requestStatus,
  error,
  isLoading,
  errors,
}) => {
  return (
    <Form className="flotilla-form-container">
      {requestStatus === RequestStatus.ERROR && error && (
        <ErrorCallout error={error} />
      )}
      <BaseTaskForm
        setFieldValue={setFieldValue}
        values={values}
        errors={errors}
      />
      <Button
        id="submit-button"
        type="submit"
        disabled={isLoading || isValid === false}
        intent={Intent.PRIMARY}
      >
        submit
      </Button>
    </Form>
  )
}

type ConnectedProps = RouteComponentProps & {
  definitionID: string
}

const Connected: React.FunctionComponent<ConnectedProps> = props => (
  <TaskContext.Consumer>
    {(ctx: TaskCtx) => {
      // Render the form if everything is ready.
      if (ctx.requestStatus === RequestStatus.READY && ctx.data) {
        const initialValues: UpdateTaskPayload = {
          env: get(ctx.data, "env", []),
          image: get(ctx.data, "image", ""),
          group_name: get(ctx.data, "group_name", ""),
          memory: get(ctx.data, "memory", 0),
          command: get(ctx.data, "command", ""),
          tags: get(ctx.data, "tags", []),
        }
        return (
          <Request<Task, { definitionID: string; data: UpdateTaskPayload }>
            requestFn={api.updateTask}
            shouldRequestOnMount={false}
            onSuccess={(data: Task) => {
              Toaster.show({
                message: `Task ${data.alias} updated successfully!`,
                intent: Intent.SUCCESS,
              })
              // Return to task page, re-request data.
              ctx.request({ definitionID: ctx.definitionID })
              props.history.push(`/tasks/${ctx.definitionID}`)
            }}
            onFailure={() => {
              Toaster.show({
                message: "An error occurred.",
                intent: Intent.DANGER,
              })
            }}
          >
            {requestProps => (
              <Formik
                initialValues={initialValues}
                validationSchema={validationSchema}
                onSubmit={data => {
                  requestProps.request({
                    data,
                    definitionID: ctx.definitionID,
                  })
                }}
              >
                {({ values, setFieldValue, isValid, errors }) => (
                  <UpdateTaskForm
                    values={values}
                    setFieldValue={setFieldValue}
                    isValid={isValid}
                    requestStatus={requestProps.requestStatus}
                    isLoading={requestProps.isLoading}
                    error={requestProps.error}
                    errors={errors}
                  />
                )}
              </Formik>
            )}
          </Request>
        )
      }

      // If there's an issue fetching task data, render an error.
      if (ctx.requestStatus === RequestStatus.ERROR && ctx.error) {
        return <ErrorCallout error={ctx.error} />
      }

      return <Spinner />
    }}
  </TaskContext.Consumer>
)

export default Connected

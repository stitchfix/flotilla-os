import * as React from "react"
import { RouteComponentProps } from "react-router-dom"
import { Button, Intent, FormGroup, Classes } from "@blueprintjs/core"
import { Formik, Form, FastField, FormikProps } from "formik"
import * as Yup from "yup"
import api from "../api"
import { CreateTaskPayload, Task } from "../types"
import Request, {
  RequestStatus,
  ChildProps as RequestChildProps,
} from "./Request"
import BaseTaskForm, {
  validationSchema as baseTaskFormValidationSchema,
} from "./BaseTaskForm"
import Toaster from "./Toaster"
import ErrorCallout from "./ErrorCallout"
import FieldError from "./FieldError"

export const validationSchema = Yup.object().shape({
  ...baseTaskFormValidationSchema,
  alias: Yup.string()
    .min(1)
    .required("Required"),
})

export type Props = Pick<
  FormikProps<CreateTaskPayload>,
  "values" | "setFieldValue" | "isValid" | "errors"
> &
  Pick<
    RequestChildProps<Task, { data: CreateTaskPayload }>,
    "requestStatus" | "error" | "isLoading"
  >

export const CreateTaskForm: React.FunctionComponent<Props> = ({
  values,
  isValid,
  setFieldValue,
  requestStatus,
  error,
  isLoading,
  errors,
}) => {
  return (
    <>
      {requestStatus === RequestStatus.ERROR && error && (
        <ErrorCallout error={error} />
      )}
      <Form className="flotilla-form-container">
        <FormGroup
          label="Alias"
          helperText="Choose a descriptive alias for this task."
        >
          <FastField className={Classes.INPUT} name="alias" />
          {errors.alias && <FieldError>{errors.alias}</FieldError>}
        </FormGroup>
        <BaseTaskForm
          setFieldValue={setFieldValue}
          values={values}
          errors={errors}
        />
        <Button
          id="submitButton"
          type="submit"
          disabled={isLoading || isValid === false}
          intent={Intent.PRIMARY}
        >
          Submit
        </Button>
      </Form>
    </>
  )
}

export type ConnectedProps = RouteComponentProps & {
  initialValues: CreateTaskPayload
  onSuccess?: (data: Task) => void
}

const Connected: React.FunctionComponent<ConnectedProps> = props => (
  <Request<Task, { data: CreateTaskPayload }>
    requestFn={api.createTask}
    shouldRequestOnMount={false}
    onSuccess={(data: Task) => {
      Toaster.show({
        message: `Task ${data.alias} created successfully!`,
        intent: Intent.SUCCESS,
      })
      props.history.push(`/tasks/${data.definition_id}`)

      if (props.onSuccess) {
        props.onSuccess(data)
      }
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
        initialValues={props.initialValues}
        validationSchema={validationSchema}
        onSubmit={data => {
          requestProps.request({ data })
        }}
      >
        {({ values, setFieldValue, isValid, errors }) => (
          <CreateTaskForm
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

Connected.defaultProps = {
  initialValues: {
    env: [],
    image: "",
    group_name: "",
    alias: "",
    memory: 1024,
    command: "",
    tags: [],
  },
}

export default Connected

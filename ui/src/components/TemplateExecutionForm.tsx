import * as React from "react"
import { Formik, Form, FastField, Field } from "formik"
import * as Yup from "yup"
import { RouteComponentProps } from "react-router-dom"
import JSONInput from "react-json-editor-ajrm"
import locale from "react-json-editor-ajrm/locale/en"
import {
  FormGroup,
  Button,
  Intent,
  Spinner,
  Classes,
  RadioGroup,
  Radio,
  Collapse,
  Colors,
} from "@blueprintjs/core"
import api from "../api"
import { TemplateExecutionRequest, Run, ExecutionEngine } from "../types"
import Request, {
  ChildProps as RequestChildProps,
  RequestStatus,
} from "./Request"
import EnvFieldArray from "./EnvFieldArray"
import ClusterSelect from "./ClusterSelect"
import { TemplateContext, TemplateCtx } from "./Template"
import Toaster from "./Toaster"
import ErrorCallout from "./ErrorCallout"
import FieldError from "./FieldError"
import NodeLifecycleSelect from "./NodeLifecycleSelect"
import * as helpers from "../helpers/runFormHelpers"
import { useSelector } from "react-redux"
import { RootState } from "../state/store"
import { getInitialValuesForTemplateExecutionForm } from "../helpers/getInitialValuesForExecutionForm"

const validationSchema = Yup.object().shape({
  owner_id: Yup.string(),
  cluster: Yup.string().required("Required"),
  memory: Yup.number()
    .required("Required")
    .min(0),
  cpu: Yup.number()
    .required("Required")
    .min(512),
  env: Yup.array().of(
    Yup.object().shape({
      name: Yup.string().required(),
      value: Yup.string().required(),
    })
  ),
  engine: Yup.string()
    .matches(/(eks|ecs)/)
    .required("A valid engine type of ecs or eks must be set."),
  node_lifecycle: Yup.string().matches(/(spot|ondemand)/),
  template_payload: Yup.object().required("Template payload is required."),
})

type Props = RequestChildProps<
  Run,
  { templateID: string; data: TemplateExecutionRequest }
> & {
  templateID: string
  initialValues: TemplateExecutionRequest
}

const TemplateExecutionForm: React.FC<Props> = ({
  initialValues,
  request,
  requestStatus,
  isLoading,
  error,
  templateID,
}) => {
  return (
    <Formik<TemplateExecutionRequest>
      isInitialValid={(values: any) =>
        validationSchema.isValidSync(values.initialValues)
      }
      initialValues={initialValues}
      validationSchema={validationSchema}
      onSubmit={data => {
        request({ templateID, data })
      }}
    >
      {({ errors, values, setFieldValue, isValid, ...rest }) => {
        const getEngine = (): ExecutionEngine => values.engine
        console.log(values)
        return (
          <Form className="flotilla-form-container">
            {requestStatus === RequestStatus.ERROR && error && (
              <ErrorCallout error={error} />
            )}
            {/* Owner ID Field */}
            <FormGroup
              label={helpers.ownerIdFieldSpec.label}
              helperText={helpers.ownerIdFieldSpec.description}
            >
              <FastField
                name={helpers.ownerIdFieldSpec.name}
                value={values.owner_id}
                className={Classes.INPUT}
              />
              {errors.owner_id && <FieldError>{errors.owner_id}</FieldError>}
            </FormGroup>
            {/* Engine Type Field */}
            <FormGroup>
              <RadioGroup
                inline
                label="Engine Type"
                onChange={(evt: React.FormEvent<HTMLInputElement>) => {
                  setFieldValue("engine", evt.currentTarget.value)

                  if (evt.currentTarget.value === ExecutionEngine.EKS) {
                    setFieldValue(
                      "cluster",
                      process.env.REACT_APP_EKS_CLUSTER_NAME || ""
                    )
                  } else if (getEngine() === ExecutionEngine.EKS) {
                    setFieldValue("cluster", "")
                  }
                }}
                selectedValue={values.engine}
              >
                <Radio label="EKS" value={ExecutionEngine.EKS} />
                <Radio label="ECS" value={ExecutionEngine.ECS} />
              </RadioGroup>
            </FormGroup>
            {/*
                Cluster Field. Note: this is a "Field" rather than a
                "FastField" as it needs to re-render when value.engine is
                updated.
            */}
            {getEngine() !== ExecutionEngine.EKS && (
              <FormGroup
                label="Cluster"
                helperText="Select a cluster for this task to execute on."
              >
                <Field
                  name="cluster"
                  component={ClusterSelect}
                  value={values.cluster}
                  onChange={(value: string) => {
                    setFieldValue("cluster", value)
                  }}
                />
                {errors.cluster && <FieldError>{errors.cluster}</FieldError>}
              </FormGroup>
            )}
            {/* CPU Field */}
            <FormGroup
              label={helpers.cpuFieldSpec.label}
              helperText={helpers.cpuFieldSpec.description}
            >
              <FastField
                type="number"
                name={helpers.cpuFieldSpec.name}
                className={Classes.INPUT}
                min="512"
              />
              {errors.cpu && <FieldError>{errors.cpu}</FieldError>}
            </FormGroup>
            {/* Memory Field */}
            <FormGroup
              label={helpers.memoryFieldSpec.label}
              helperText={helpers.memoryFieldSpec.description}
            >
              <FastField
                type="number"
                name={helpers.memoryFieldSpec.name}
                className={Classes.INPUT}
              />
              {errors.memory && <FieldError>{errors.memory}</FieldError>}
            </FormGroup>
            <FormGroup
              label={helpers.nodeLifecycleFieldSpec.label}
              helperText={helpers.nodeLifecycleFieldSpec.description}
            >
              <Field
                name={helpers.nodeLifecycleFieldSpec.name}
                component={NodeLifecycleSelect}
                value={values.node_lifecycle}
                onChange={(value: string) => {
                  setFieldValue(helpers.nodeLifecycleFieldSpec.name, value)
                }}
                isDisabled={getEngine() !== ExecutionEngine.EKS}
              />
              {errors.node_lifecycle && (
                <FieldError>{errors.node_lifecycle}</FieldError>
              )}
            </FormGroup>
            <FormGroup label="Template Payload">
              <FastField
                className={Classes.CODE}
                component={JSONInput}
                name="template_payload"
                placeholder={values.template_payload}
                onChange={({ jsObject }: any) => {
                  setFieldValue("template_payload", jsObject)
                }}
                colors={{
                  background: Colors.DARK_GRAY2,
                }}
                width={600}
                height={400}
                style={{
                  body: {
                    fontSize: "13px",
                  },
                }}
                locale={locale}
              />
              {errors.template_payload && (
                <FieldError>{errors.template_payload}</FieldError>
              )}
            </FormGroup>
            <EnvFieldArray />
            <Button
              intent={Intent.PRIMARY}
              type="submit"
              disabled={isLoading || isValid === false}
              style={{ marginTop: 24 }}
              large
            >
              Submit
            </Button>
          </Form>
        )
      }}
    </Formik>
  )
}

const Connected: React.FunctionComponent<RouteComponentProps> = ({
  location,
  history,
}) => {
  const { settings } = useSelector((s: RootState) => s.settings)
  return (
    <Request<Run, { templateID: string; data: TemplateExecutionRequest }>
      requestFn={api.runTemplate}
      shouldRequestOnMount={false}
      onSuccess={(data: Run) => {
        Toaster.show({
          message: `Run ${data.run_id} submitted successfully!`,
          intent: Intent.SUCCESS,
        })
        history.push(`/runs/${data.run_id}`)
      }}
      onFailure={() => {
        Toaster.show({
          message: "An error occurred.",
          intent: Intent.DANGER,
        })
      }}
    >
      {requestProps => (
        <TemplateContext.Consumer>
          {(ctx: TemplateCtx) => {
            switch (ctx.requestStatus) {
              case RequestStatus.ERROR:
                return <ErrorCallout error={ctx.error} />
              case RequestStatus.READY:
                if (ctx.data) {
                  const initialValues: TemplateExecutionRequest = getInitialValuesForTemplateExecutionForm(
                    ctx.data,
                    location.state
                  )
                  return (
                    <TemplateExecutionForm
                      templateID={ctx.templateID}
                      initialValues={initialValues}
                      {...requestProps}
                    />
                  )
                }
                break
              case RequestStatus.NOT_READY:
              default:
                return <Spinner />
            }
          }}
        </TemplateContext.Consumer>
      )}
    </Request>
  )
}

export default Connected

import * as React from "react"
import { Formik, FormikProps, Form } from "formik"
import { get, Omit } from "lodash"
import {
  IFlotillaCreateTaskPayload,
  IReactSelectOption,
  flotillaUIRequestStates,
  IFlotillaEditTaskPayload,
  IFlotillaAPIError,
  IFlotillaTaskDefinition,
} from "../../types"
import TaskFormNavigation from "./TaskFormNavigation"
import FormikKVField from "../Field/FormikKVField"
import View from "../styled/View"
import StyledForm from "../styled/Form"
import api from "../../api"
import * as TaskFormFields from "./Fields"
import { CreateTaskYupSchema } from "./validation"
import Request, { IChildProps as IRequestChildProps } from "../Request/Request"
import Loader from "../styled/Loader"

export type TaskFormPayload =
  | IFlotillaCreateTaskPayload
  | IFlotillaEditTaskPayload

export interface IProps {
  defaultValues: TaskFormPayload
  groupOptions: IReactSelectOption[]
  tagOptions: IReactSelectOption[]
  title: string
  submitFn: (values: TaskFormPayload) => Promise<IFlotillaTaskDefinition>
  onSuccess?: (definition: IFlotillaTaskDefinition) => void
  onFail?: (error: IFlotillaAPIError) => void
  validateSchema: any
}

interface IState {
  inFlight: boolean
  error: any
}

export class BaseTaskForm extends React.PureComponent<IProps, IState> {
  static defaultProps: Partial<IProps> = {
    validateSchema: CreateTaskYupSchema,
  }
  state = {
    inFlight: false,
    error: false,
  }

  handleSubmit = (values: TaskFormPayload) => {
    const { submitFn, onSuccess, onFail } = this.props

    this.setState({ inFlight: true })

    submitFn(values)
      .then((res: IFlotillaTaskDefinition) => {
        this.setState({ inFlight: false })

        if (onSuccess) onSuccess(res)
      })
      .catch((error: IFlotillaAPIError) => {
        this.setState({ inFlight: false, error })

        if (onFail) onFail(error)
      })
  }

  render() {
    const { defaultValues, title, groupOptions, tagOptions } = this.props
    const { inFlight } = this.state

    return (
      <Formik
        initialValues={defaultValues}
        onSubmit={this.handleSubmit}
        validationSchema={CreateTaskYupSchema}
      >
        {(formikProps: FormikProps<TaskFormPayload>) => {
          return (
            <Form>
              <View>
                <TaskFormNavigation
                  isSubmitDisabled={formikProps.isValid !== true}
                  inFlight={inFlight}
                />
                <StyledForm title={title}>
                  <TaskFormFields.AliasField />
                  <TaskFormFields.GroupNameField
                    onChange={(value: string) => {
                      formikProps.setFieldValue("group_name", value)
                    }}
                    value={formikProps.values.group_name}
                    options={groupOptions}
                  />
                  <TaskFormFields.ImageField />
                  <TaskFormFields.CommandField />
                  <TaskFormFields.MemoryField />
                  <TaskFormFields.TagsField
                    onChange={(value: string[]) => {
                      formikProps.setFieldValue("tags", value)
                    }}
                    value={get(formikProps, ["values", "tags"], [])}
                    options={tagOptions}
                  />
                  <FormikKVField
                    name="env"
                    value={formikProps.values.env}
                    description="Environment variables that can be adjusted during execution."
                    isKeyRequired
                    isValueRequired={false}
                    label="Environment Variables"
                    setFieldValue={formikProps.setFieldValue}
                  />
                </StyledForm>
              </View>
            </Form>
          )
        }}
      </Formik>
    )
  }
}

const BaseTaskFormWithSelectOptions: React.SFC<
  Omit<IProps, "groupOptions" | "tagOptions">
> = props => (
  <Request
    shouldRequestOnMount
    requestFn={[api.getGroups, api.getTags]}
    initialRequestArgs={[]}
  >
    {(requestProps: IRequestChildProps) => {
      if (requestProps.requestState === flotillaUIRequestStates.READY) {
        return (
          <BaseTaskForm
            {...props}
            groupOptions={requestProps.data[0]}
            tagOptions={requestProps.data[1]}
          />
        )
      }

      return <Loader />
    }}
  </Request>
)

export default BaseTaskFormWithSelectOptions

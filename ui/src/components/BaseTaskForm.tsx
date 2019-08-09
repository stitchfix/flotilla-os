import * as React from "react"
import { FormGroup, Classes } from "@blueprintjs/core"
import { FastField, FormikProps } from "formik"
import * as Yup from "yup"
import GroupNameSelect from "./GroupNameSelect"
import TagsSelect from "./TagsSelect"
import EnvFieldArray from "./EnvFieldArray"
import FieldError from "./FieldError"

export const validationSchema = {
  env: Yup.array().of(
    Yup.object().shape({
      name: Yup.string().required(),
      value: Yup.string().required(),
    })
  ),
  image: Yup.string()
    .min(1)
    .required("Required"),
  group_name: Yup.string()
    .min(1)
    .required("Required"),
  memory: Yup.number()
    .required("Required")
    .min(0),
  command: Yup.string()
    .min(1)
    .required("Required"),
  tags: Yup.array().of(Yup.string()),
}

export type Props = Pick<
  FormikProps<any>,
  "values" | "setFieldValue" | "errors"
>
export const BaseTaskForm: React.FunctionComponent<Props> = ({
  values,
  setFieldValue,
  errors,
}) => (
  <>
    <FormGroup
      label="Group Name"
      helperText="Create a new group name or select an existing one to help searching for this task in the future."
    >
      <FastField
        name="group_name"
        component={GroupNameSelect}
        value={values.group_name}
        onChange={(value: string) => {
          setFieldValue("group_name", value)
        }}
      />
      {errors.group_name && <FieldError>{errors.group_name}</FieldError>}
    </FormGroup>
    <FormGroup
      label="Docker Image"
      helperText="The full URL of the Docker image and tag."
    >
      <FastField name="image" className={Classes.INPUT} />
      {errors.image && <FieldError>{errors.image}</FieldError>}
    </FormGroup>
    <FormGroup
      label="Command"
      helperText="The command for this task to execute."
    >
      <FastField
        className={`${Classes.INPUT} ${Classes.CODE}`}
        component="textarea"
        name="command"
        rows={14}
      />
      {errors.command && <FieldError>{errors.command}</FieldError>}
    </FormGroup>
    <FormGroup
      label="Memory (MB)"
      helperText="The amount of memory (MB) this task needs."
    >
      <FastField type="number" name="memory" className={Classes.INPUT} />
      {errors.memory && <FieldError>{errors.memory}</FieldError>}
    </FormGroup>
    <FormGroup label="Tags">
      <FastField
        name="tags"
        component={TagsSelect}
        value={values.tags}
        onChange={(value: string[]) => {
          setFieldValue("tags", value)
        }}
      />
      {errors.tags && <FieldError>{errors.tags}</FieldError>}
    </FormGroup>
    <EnvFieldArray />
  </>
)

export default BaseTaskForm

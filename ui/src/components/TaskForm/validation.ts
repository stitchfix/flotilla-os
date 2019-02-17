import * as Yup from "yup"

const shared = {
  command: Yup.string()
    .min(1, "")
    .required("Required"),
  memory: Yup.number()
    .min(1, "")
    .required("Required"),
  image: Yup.string()
    .min(1, "")
    .required("Required"),
  group_name: Yup.string()
    .min(1, "")
    .required("Required"),
  tags: Yup.array().of(
    Yup.string()
      .min(1, "")
      .required("Required")
  ),
  env: Yup.array().of(
    Yup.object().shape({
      name: Yup.string()
        .min(1, "")
        .required("Required"),
      value: Yup.string(),
    })
  ),
}

export const CreateTaskYupSchema = Yup.object().shape({
  alias: Yup.string()
    .min(1, "")
    .required("Required"),
  ...shared,
})

export const UpdateTaskYupSchema = Yup.object().shape(shared)

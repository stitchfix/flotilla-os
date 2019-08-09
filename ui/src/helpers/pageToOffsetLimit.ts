const pageToOffsetLimit = ({
  page,
  limit,
}: {
  page: number
  limit: number
}) => ({
  offset: (page - 1) * limit,
  limit,
})

export default pageToOffsetLimit

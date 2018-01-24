describe("TaskHistory", () => {
  describe("Lifecycle Methods", () => {
    describe("componentDidMount", () => {
      it(
        "calls `props.updateQuery` with the default query if `props.query` is empty"
      )
      it("calls `this.fetch` if a `props.query` is not empty")
    })
    describe("componentWillReceiveProps", () => {
      it(
        "calls `this.fetch` with the correct query and url if the query or definition ID changes"
      )
    })
  })
  describe("render", () => {
    it("renders a <Loader /> if props.isLoading")
    it("renders an <ErrorCard /> if props.error")
    it("renders a helpful message if no data is found")
    it("renders the data if present")
  })
})

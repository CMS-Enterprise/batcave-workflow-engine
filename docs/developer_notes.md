# Developer Notes

## Tidy First

By: Kent Beck

"Tidy First?" suggests the following:

- There isn’t a single way to do things, there are things that make sense in context, and you know your context
- There are many distinct ways to tidy code, which make code easier to work with: guard clauses, removing dead code, normalizing symmetries, and so on
- Tidying and logic changes are different types of work, and should be done in distinct pull requests
- This speeds up pull request review, and on high-cohesion teams tidying commits shouldn’t require code review at all
- Tidying should be done in small amounts, not large amounts
- Tidying is usually best to do before changing application logic, to the extent that it reduces the cost of making the logical change
- It’s also OK to tidy after your change, later when you have time, or even never (for code that doesn’t change much)
- Coupling is really bad for maintainable code

## Effective Go

link: [Effective Go](https://go.dev/doc/effective_go)

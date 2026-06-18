---
name: tdd
description: Expert guidance for applying Test-Driven Development (TDD) and Red-Green-Refactor workflows. You MUST use this skill when implementing new functions, methods, packages, or features, or when refactoring existing code to ensure safety. This skill is critical when the user mentions TDD, test-first, write tests, unit test, write test cases, or whenever adding new code. Do not write implementation code without first writing a failing test case in this skill's RED phase.
---

# Test-Driven Development (TDD) Skill

This skill enforces the Test-Driven Development (TDD) discipline. TDD is a software development process relying on software requirements being converted to test cases before software is fully developed, and tracking all software development by repeatedly testing the software against all test cases.

## Core Cycle: Red-Green-Refactor

You MUST strictly follow this three-phase cycle:

### 1. RED Phase (Write a Failing Test First)
- **Goal**: Write a test that specifies a desired behavior, compiles (with stubs if necessary), and fails.
- **Actions**:
  1. **Identify the requirements**: Understand what the new function, method, struct, or API needs to do.
  2. **Create the test file & case**: Write unit tests in a test file (e.g., `feature_test.go` or equivalent). Focus on testing the public interface, defining input and expected output.
  3. **Create the Implementation Stub**: If the code does not compile because the type/function does not exist, write the minimal empty definition (a stub) in the implementation file (e.g., `func CalculateArea(width, height float64) float64 { return 0.0 }`).
  4. **Run the test**: Run the test suite (e.g., `go test`). Confirm that:
     - The tests compile successfully.
     - The new test case fails for logical reasons (e.g., actual value `0.0` does not match expected value `12.5`).
     - **CRITICAL**: Never proceed to implementation without seeing the test fail first!

### 2. GREEN Phase (Make it Pass)
- **Goal**: Write the simplest code that makes the failing test pass.
- **Actions**:
  1. Write the minimal logic required to make the test pass.
  2. Avoid implementing additional, unrequested features or optimizing the code at this stage. Quick-and-dirty implementation is acceptable here if it gets the tests green.
  3. Run the test suite. If tests fail, iterate on the implementation until they pass.
  4. Once the tests are green, verify that all other existing tests in the suite are also green.

### 3. REFACTOR Phase (Clean the Code)
- **Goal**: Clean up the design and implementation while keeping the tests green.
- **Actions**:
  1. **Review implementation and tests**: Look for code duplication, unclear naming, hardcoded values, complex loops, lack of comments, or style guide violations.
  2. **Refactor code**: Improve the code structure without changing its external behavior.
  3. **Refactor tests**: Ensure the test code is also clean, readable, and easy to maintain.
  4. **Verify**: Run the test suite after each minor refactoring step to ensure you didn't break anything. The tests must remain green.

---

## Refactoring Existing Code

When using this skill to refactor existing code, follow this procedure to ensure safety:

1. **Verify Baseline (Establish Safety)**: 
   - Before making any code changes, run the existing unit tests. They MUST be green.
   - If unit tests for the code to be refactored do not exist, write them first! Do not refactor without tests. This is your safety net.
2. **Refactor in Tiny Steps**:
   - Make small, incremental changes (e.g., extract a variable, rename a function, simplify a conditional).
   - Do NOT try to rewrite the entire module in one go.
3. **Keep Tests Green**:
   - Run the tests after every single change. If they fail, immediately revert the last change or fix it.
4. **Clean Test Code**:
   - Refactoring applies to test files too. Keep tests readable and remove redundant/flaky checks.

---

## Common Mistakes to Avoid
- **Writing implementation before tests**: This defeats the entire purpose of TDD.
- **Not verifying the Red phase**: If a test passes immediately, either the feature is already implemented, or the test is invalid (e.g., asserting `nil != nil` or missing assertions).
- **Over-engineering in the Green phase**: Do not try to solve future problems or build excessive abstractions before they are tested.
- **Skipping Refactoring**: Getting the code working (green) is not the end. You must clean it up to prevent technical debt.

## Language Specific Guides (Go/Golang)
- Write tests using the standard library `testing` package or `github.com/stretchr/testify` for assertions.
- Test files must have the suffix `_test.go` and be in the same package (or `package_test` for external testing).
- Keep tests fast by avoiding unnecessary network calls or database writes unless using mock interfaces.

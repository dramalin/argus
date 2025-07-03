// Package tasks provides functionality for scheduling and managing system maintenance tasks
package tasks

/*
Task System Test Notes

This file documents the testing approach for the task management system and explains
known issues that need to be addressed in future refactoring efforts.

1. Import Cycle Issue

   An import cycle exists between the tasks package and the repository package:
   - tasks/scheduler.go imports tasks/repository
   - tasks/repository/repository.go imports tasks for model types

   This circular dependency prevents running all tests together and requires running
   tests within specific package contexts.

2. Suggested Future Refactoring

   To resolve this issue properly, the codebase should be refactored to either:
   a) Create a shared models package containing all shared types
   b) Move repository interface into the tasks package and only reference it there
   c) Use dependency injection with interfaces defined in the consumer packages

3. Current Testing Approach

   Tests have been written for individual components:
   - types_test.go: Tests for the basic model types
   - runner_test.go: Tests for the task runners
   - repository_test.go: Tests for the task repository implementation
   - scheduler_test.go: Tests for the task scheduler

   These tests must be run within their respective package contexts.

4. Test Coverage

   Current test coverage includes:
   - Task model validation and basic functionality
   - Task runner creation and execution for all task types
   - Repository file operations and data persistence
   - Scheduler task registration and execution

5. Testing Instructions

   To test individual components, run:
   - cd internal/tasks && go test -v ./... -run TestTaskConfig_Validate
   - cd internal/tasks && go test -v ./... -run TestNewTaskRunner
   - cd internal/tasks/repository && go test -v

6. Integration Tests

   Integration tests for the entire task system should be added once the
   import cycle issue is resolved, focusing on:
   - End-to-end task scheduling
   - Task execution results and error handling
   - API integration with task system
*/

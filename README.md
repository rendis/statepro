# StatePro: Advanced State Machine Handling in Golang
[![Go Report Card](https://goreportcard.com/badge/github.com/rendis/statepro)](https://goreportcard.com/report/github.com/rendis/statepro)
[![Go Reference](https://pkg.go.dev/badge/github.com/rendis/statepro.svg)](https://pkg.go.dev/github.com/rendis/statepro)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/release/rendis/statepro.svg?style=flat-square)](https://github.com/rendis/statepro/releases)


<p align="center">
    <img src="documentation%2Fassets%2Fdog-machine-diagram.png" alt="golang state machine inspired by xstate" width="800">
</p>

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
2. [Components of a State Machine](#components-of-a-state-machine)
3. [Initializing the State Machine](#initializing-the-state-machine)
    - [JSON Definition of the State Machine](#json-definition-of-the-state-machine)
    - [Context](#context)
    - [Definition Registry](#definition-registry)
4. [Initialization](#initialization)
5. [Getting an instance of the state machine](#getting-an-instance-of-the-state-machine)
6. [ProMachine](#promachine)
7. [Creating Action, Invocation, Guard and Context Handlers](#creating-action-invocation-guard-and-context-handlers)
    - [Event](#event)
    - [ActionTool](#actiontool)
    - [Method Definitions](#method-definitions)
8. [Examples](#examples)

## Introduction

StatePro is a Golang library for handling Finite State Machines, designed to optimize state management in microservices.
Inspired by [XState](https://xstate.js.org/) but focused on backend development,
the JSON representation of the State Machine is compatible with XState's [visual creator](https://stately.ai/),
facilitating its design and visualization.

Despite similarities with XState, StatePro has its own set of features, primarily to ensure efficiency and adaptability in
state management in microservices. This is because some XState features, although useful for frontend development, do not always
translate effectively to the microservices environment, which can lead to unnecessary complexity.

## Installation

To install StatePro, use the following command:

```bash
go get github.com/rendis/statepro
```

## Components of a State Machine

1. **State Machine**: Defines the overall behavior of the system and is composed of states, transitions, events,
   actions, guards, and invocations.

2. **State**: A state represents a particular stage in the lifecycle of a state machine.

3. **Action**: Actions are behaviors that are performed when an event or transition occurs. Actions are synchronous and are executed in the order in which they are defined.

4. **Guard**: Guards are conditions that must be met for a transition to occur.

5. **Invocation**: Invocations are asynchronous tasks that are performed upon entering a state. Unlike other
   implementations, in StatePro, the result of an invocation (success or failure) does not affect the behavior of the
   state machine.

6. **Transition**: A transition is a change from one state to another in response to an event.

7. **Event**: Events are inputs that trigger transitions between states.

8. **Context**: The context is an object associated with the state machine that can be modified within actions and
   queried by the other components.

## Initializing the State Machine

To initialize a state machine in StatePro, three components are needed: the JSON definition of the state machine, the context, and the definition registry.

### JSON Definition of the State Machine

The JSON will contain the definition of the state machine. This can be created in [Stately](https://stately.ai/) and imported into your project.
It is important that each JSON definition you load into your system has a unique ID. This ID will be used later to associate the JSON with the definition registry.

The location of the state machine definition JSONs is configurable. By default, StatePro will look in the `statepro.yml` file at the root of the project.
You can change the location and filename of the configuration file using the `SetDefinitionPath` method of the `statepro` package:

```go
import "github.com/rendis/statepro"

statepro.SetDefinitionPath("path/to/your/prop.yml")
```

Within the configuration file, you must define the location of the state machine definition JSON files. For example:

```yml
statepro:
  file-prefix: '<prefix>'
  paths:
    - '<rute1>'
    - '<rute2>'
    ...
```

- `file-prefix`: specifies the prefix that file names must have to be considered as definition files. Only files whose name begins with this prefix will be processed.
- `paths`: is a list of paths that specify the directories and/or files that should be searched to find the state machine definitions.

### Context

The context is a struct that will be linked to the behavior of the state machine.

### Definition Registry

The definition registry is a struct that will contain the action, invocation, guard methods, and context handling methods that will be used in the state machine.
This struct should have the type of the `context` as a generic.

This struct should implement the `MachineRegistryDefinitions` interface, which is used to get the ID of the state machine.
This ID should be the same as the ID of the definition JSON that you want to associate it with.


For example:

```go
type Context struct {
    name string
    state ContextState
    ...
}

type ContextMachineDefinitions[ContextType Context] struct {}

// Implementation of the GetMachineTemplateId method of the MachineRegistryDefinitions interface
func (cmd *ContextMachineDefinitions[ContextType]) GetMachineTemplateId() string {
    return "MACHINE_ID"
}

// Implementation of action, invocation, guard methods, and context handling methods
```

Once you have these three components (JSON definition, context, and definition registry), you can initialize your state machine
and start using it to manage the flow of your application.

## Initialization

Before initializing `statepro`, you must have registered **all** the state machine definitions using the `AddMachine` method of the `statepro` package.
To initialize a state machine, you should use the `InitMachines` method of the `statepro` package.

```go
var definition1 = ContextMachineDefinitions[Context]{}
var definition2 = ContextMachineDefinitions2[Context2]{}
var definition3 = ContextMachineDefinitions3[Context3]{}
...

// register definitions
var machineId1 = statepro.AddMachine[Context](definition1)
var machineId2 = statepro.AddMachine[Context2](definition2)
var machineId3 = statepro.AddMachine[Context3](definition3)
...

// InitMachines should be called after registering all definitions
statepro.InitMachines()
```

The return value of `AddMachine` is the ID associated with the state machine created from the definition registry `ContextMachineDefinitions[Context]`.
This ID will be unique and will be used to get an instance of the state machine.
> Note: The ID returned by `AddMachine` is not the same as the ID of the JSON definition. It is a unique ID generated for each state machine.

## Getting an instance of the state machine

To get a state machine, you should use the `GetMachine` method of the `statepro` package, which returns an object of type `ProMachine`.
This is an interface that defines the methods that allow you to interact with the state machine.

```go
var context = &Context{}

var contextMachine, err = statepro.GetMachine[Context](machineId1, context)
```

To obtain a state machine, in addition to its ID (`machineId1`), an instance of the context must be provided.
If it is `nil`, an instance of the context will be attempted to be obtained through the `ContextFromSource` method of the associated definition registry.
If an instance of the context does not exist in the `ContextFromSource` method, an error will be returned.


## ProMachine

StatePro defines a `ProMachine` interface with several methods that allow you to interact with the state machine:

```go
type ProMachine[ContextType any] interface {
    PlaceOn(stateName string) error
    StartOn(stateName string) TransitionResponse
    StartOnWithEvent(stateName string, event Event) TransitionResponse
    SendEvent(event Event) TransitionResponse
    GetNextEvents() []string
    GetState() string
    IsFinalState() bool
    GetContext() ContextType
    CallContextToSource() error
}
``` 

- `PlaceOn(stateName string) error`: places the state machine in a specific state, without executing entry actions.

- `StartOn(stateName string) TransitionResponse`: places the state machine in a specific state and executes the entry actions.

- `StartOnWithEvent(stateName string, event Event) TransitionResponse`: is similar to `StartOn`, but also sends an event to the state machine.

- `SendEvent(event Event) TransitionResponse`: allows sending an event to the state machine, which can trigger transitions and actions.

- `GetNextEvents() []string`: returns a list of the names of events that can be sent from the current state.

- `GetState() string`: returns the name of the current state.

- `IsFinalState() bool`: checks if the current state is a final state.

- `GetContext() ContextType`: allows getting the value of the context associated with the state machine.

- `CallContextToSource() error`: allows calling the 'ContextToSource' method, if it exists, in the definition registry of the state machine.

The response structure `TransitionResponse` contains information about the transition(s) that occurred.
```go
type TransitionResponse interface {
    GetLastEvent() Event
    Error() error
}
```

- `GetLastEvent() Event`: returns the last event that was sent to the state machine.
- `Error() error`: returns an error if one occurred during the transition.

Regarding the features related to **states** and **transitions**, StatePro allows defining specific behaviors that are defined in the design of the state machine:

#### About a State

- **Execute Entry Actions**: When entering a state, specific actions can be executed.

- **Execute Exit Actions**: When leaving a state, specific actions can be executed.

- **Execute Invocations**: When entering a state, asynchronous tasks can be executed. Although invocations are
  asynchronous, in StatePro, their result (success or failure) does not affect the decision-making of the state machine.
  For example, when entering a state, you might want to asynchronously send an event to a message queue. If the
  invocation fails, the state machine will not be affected and will continue with its normal execution.

#### About Transitions

- **Execute Transition Actions**: During a transition, specific actions can be executed.

- **Execute Guards**: Before performing a transition, conditions can be evaluated that determine whether the transition
  should be performed or not.

- **Execute Guard Actions**: While evaluating a guard, specific actions can be executed.

## Creating Action, Invocation, Guard and Context Handlers

To work with StatePro, it is important to understand how to create and use actions, invocations, and guards. These
methods are essential to defining the behavior of the state machine. Before delving into how to define
these methods, we'll explain two essential elements: `Event` and `ActionTool`.

### Event

The `Event` object is the basic unit of communication between the states of the machine. An `Event` can contain a
`Data` value, which can be used to pass information from one state to another. Here's its definition:

```go
type Event interface {
    GetName() string
    GetFrom() string
    HasData() bool
    GetData() any
    GetDataAsMap() (map[string] any, error)
    GetErr() error                         
    GetEvtType() EventType
    ToBuilder() EventBuilder
}
```

To build an event, you use `EventBuilder`:

```go
type EventBuilder interface {
    WithData(data any) EventBuilder
    WithErr(err error) EventBuilder
    WithType(eventType EventType) EventBuilder
    Build() Event
}
```

For example, building an event would look like this:

```go
import "github.com/rendis/statepro/piece"

var evt := piece.BuildEvent("EVENT_NAME").Build()
```

`ActionTool` is an object used to interact with the state machine from within an action. It has the following definition:

```go
type ActionTool[ContextType any] interface {
    Assign(context ContextType)
    Send(event Event)          
    Propagate(event Event)
}
```

- `Assign(context ContextType)`: This method allows assigning a new context to the state machine. The context is useful for sharing data between different states and transitions.

- `Send(event Event)`: This method allows sending an event to the state machine. The event will be processed by the current state of the machine.

- `Propagate(event Event)`: This method allows propagating an event with new data and errors throughout the operations following the current action. It is useful for transmitting additional information or errors that occur during the execution of an action.

### Method Definitions

The **action**, **guard**, and **invoke** methods must have the same name as the component they are intended to associate with in the state machine's JSON. 
It's important to note that the names in the JSON are case-insensitive. This means that if we have `doSomething`, `DoSomething`, and `dosomething` in the JSON, 
for the state machine engine, **_these three variants will be considered the same_**. Consequently, their counterpart in the code should be named exactly as `DoSomething`. 
This naming standard is crucial for maintaining consistency between the state machine's JSON and its implementation in the code.

Theses methods must be defined in the `MachineRegistryDefinitions` as follows:

```go
// Action
func(cmd *ContextMachineDefinitions[ContextType]) ActionName(contextValue Context, evt Event, actTool ActionTool[ContextType]) error {...}

// Guard
func(cmd *ContextMachineDefinitions[ContextType]) GuardName(contextValue  Context, evt Event) (bool, error) {...}

// Invocation
func (cmd *ContextMachineDefinitions[ContextType]) InvocationName(contextValue  Context, evt Event) {...}
}
```

In addition, within the definition of state machine records, two methods can be defined for getting and saving the context. These methods are useful for centralizing the logic in one place.

To define the context retrieval method, the `ContextFromSource` function must be implemented in the **StateMachineRegistry** as follows:

```go
// Context Retrieval
func (cmd *ContextMachineDefinitions[ContextType]) ContextFromSource(params ... any) (Context, error) {
    // obtain parameters from params (params[0], params[1], etc)
    // context retrieval logic
    return context, nil
}
```
And to define the context saving method, the `ContextToSource` function must be implemented in the **StateMachineRegistry** as follows:

```go
// Context Saving
func (cmd *ContextMachineDefinitions[ContextType]) ContextToSource(context Context) error {
    // context saving logic
    return nil
}
```

## Examples

In order to help you get a better understanding of how to use StatePro, we've created several examples that demonstrate its various features. These examples are organized into different branches, each focusing on a specific aspect of StatePro.

Here's a brief description of what each branch covers:

1. **01-example-basic**: This branch contains a basic example of how to use StatePro. It's the perfect starting point if you're new to the library.

2. **02-read-basic-example**: In this branch, you'll find examples of how to read the state of a state machine and how to react to changes.

3. **03-write-basic-example**: This branch focuses on writing to the state machine. You'll learn how to trigger events and cause state transitions.

4. **04-events-example**: This branch goes deeper into event handling with StatePro. It demonstrates how to define and use custom events.

5. **05-invocations-services**: In this branch, we cover the topic of invocations. You'll learn how to define and use asynchronous tasks that get executed when entering a state.

6. **06-context-handlers**: This branch focuses on context handlers. You'll learn how to define and use methods for getting and saving the context of a state machine.

To view the examples, simply switch to the respective branch. Remember, these examples are intended to be a learning resource. Feel free to modify and experiment with them as you become more comfortable with StatePro.

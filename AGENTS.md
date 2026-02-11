
# Agents in this Project – Python to Go Rewrite

## 1. Purpose of the Agent System

This project aims to rewrite an existing Python codebase into Go while:

- preserving functional behavior
- improving performance and reliability
- documenting architecture and decisions
- tracking risks and migration issues

The agents collaborate to transform raw ideas and problems into
a structured Go implementation.

---

## 2. Source of Truth

All reasoning originates from:

    PH/  – Prompt History (immutable)

Derived artifacts are created in:

- /requirements – WHAT must be preserved or changed
- /arch         – HOW the Go system will be structured
- /tasks        – Concrete migration steps
- /bug-tracker  – Investigations and migration problems
- /docs         – Explanations for developers

Direction of truth is one-way:

    PH → Requirements → Architecture → Tasks  
                  ↘ Problem Investigations ↗

No agent may modify PH files.

---

## 3. Agent Roles

### Agent 1 – Prompt History Tracker

**Role:** Memory & audit

- Records all prompts verbatim into PH/
- Hourly rotation
- Append-only

**Why for migration:**  
Preserves original intent of the Python system and all design discussions.

---

### Agent 2 – Requirements Tracker

**Role:** Product & behavior guardian

- Extracts functional behavior of the Python system
- Identifies constraints that must exist in Go
- Separates:
  - behavior to keep
  - behavior to improve
  - behavior to remove

**Key questions**
- What must the Go version do exactly like Python?
- Which side effects are accidental?
- Which contracts are implicit?

**Output**
- /requirements/*.md

---

### Agent 3 – Architecture & Implementation Agent

**Role:** Systems architect

- Designs Go structure based on requirements
- Maps Python concepts to Go equivalents:

| Python Concept | Go Target |
|----------------|-----------|
| dynamic types  | static types |
| classes        | structs + interfaces |
| exceptions     | errors |
| threads/async  | goroutines |
| monkey patch   | DI patterns |

**Creates**

- /arch/component-map.md  
- /arch/python-to-go-mapping.md  
- /arch/data-models.md  
- /arch/interfaces.md

**Responsibilities**

- Define module boundaries  
- Define concurrency model  
- Define error strategy  
- Define test approach

---

### Agent 4 – Task Planner

**Role:** Migration executor

- Breaks rewrite into safe increments:

1. Identify Python module  
2. Define Go interface  
3. Create tests  
4. Implement Go version  
5. Parity validation

**Rules**

- No task > 2 days  
- Each task has:
  - acceptance criteria
  - parity test
  - rollback path

**Output**
- /tasks/migration-YYYY-MM-DD.md

---

### Agent 5 – Problem Analysis & Insight Tracker (JSON)

**Role:** Investigator during migration

Used when:

- Go behavior differs from Python  
- performance regressions  
- semantic mismatches

**Produces**

    ./bug-tracker/BUG-<id>.json

including:

- hypotheses  
- comparison Python vs Go  
- experiments  
- test proposals

**Central question**

> Why does Go behave differently from Python?

---

## 4. Migration Workflow

### Step A – Capture Truth
Prompt Tracker records all context about the Python system.

### Step B – Extract Requirements
Requirements Agent defines:

- external contracts  
- data formats  
- timing expectations  
- error behavior

### Step C – Design Go Architecture
Architecture Agent:

- creates type system  
- defines packages  
- selects libraries  
- defines concurrency

### Step D – Plan Rewrite
Task Planner creates incremental path:

Python module → Go interface → Tests → Implementation

### Step E – Investigate Divergence
Problem Agent compares:

- Python output  
- Go output  
- edge cases

---

## 5. Quality Gates

The Go rewrite is accepted when:

1. All requirements have Go equivalents  
2. Parity tests pass  
3. Performance ≥ Python baseline  
4. Error semantics documented  
5. No open critical bugs

---

## 6. Folder Contracts

### /requirements
- behavior contracts
- invariants
- compatibility rules

### /arch
- python-to-go mapping
- package design
- concurrency model

### /tasks
- migration steps
- test plans

### /bug-tracker
- investigations
- experiments
- comparisons

### /docs
- onboarding for Go devs
- rationale

---

## 7. Governance Rules

- No agent edits another agent’s files  
- PH is immutable  
- Every decision links to PH  
- Behavior parity first, improvements second

---

## 8. Schedule

- Tracker: hourly  
- Requirements: on new PH  
- Architecture: on new requirements  
- Tasks: after architecture  
- Problem: on demand

---

## 9. Success Definition

The agents succeed when:

- The Go system can replace Python in production  
- All behaviors are explained  
- Differences are intentional  
- Knowledge is documented

---

## 10. Version

AGENTS spec: 2.0 – Go Migration Edition


⸻

How This Helps You Day-to-Day

This file gives you:

1) A migration compass
	•	No random rewriting
	•	Behavior-first
	•	Parity before optimization

2) Clear thinking tools
	•	Python → Go concept map
	•	structured investigations
	•	experiment logs

3) Safety net
	•	every divergence documented
	•	reversible decisions
	•	knowledge preserved

⸻
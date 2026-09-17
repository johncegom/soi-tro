# Idea Inbox

This file captures early product ideas before they are ready to become proposed
work. Keep entries short. An idea does not authorize implementation and does
not need a Definition of Done, Test Plan, or Program design yet.

## Lifecycle

```text
Idea inbox -> LEDGER.md (Proposed) -> task document -> implementation
```

Use one of these statuses:

- `New`: captured but not evaluated.
- `Exploring`: questions, value, or feasibility are being investigated.
- `Promoted`: accepted into `docs/LEDGER.md` with a linked task document.
- `Parked`: worth retaining, but not being considered now.
- `Rejected`: intentionally declined; keep the reason so it is not repeatedly
  proposed.

When promoting an idea, add links in both directions between this file and the
new task document. Create a separate idea folder only when supporting research
or artifacts make a single entry unwieldy.

## IDEA-001 - Personal deal-breaker filter

- Status: `Promoted`
- Task: [013 - Add personal deal-breaker filtering](tasks/013-personal-deal-breaker-filter.md)
- Problem: Users spend time reviewing or contacting rentals that violate a
  non-negotiable requirement.
- Who benefits: Renters with firm limits on budget, move-in cash, pets,
  parking, elevator access, floor, or missing information.
- Idea: Let users save personal requirements, then classify each analyzed
  listing as shortlist, needs checking, or reject with field-level reasons.
- Inverse check: Never turn unknown data into a pass or hide a hard failure
  behind an opaque match score.
- Open questions:
  - Which requirements are hard filters and which are preferences?
  - Should profiles support more than one renter or search scenario?
- Expected value: Fewer unsuitable contacts and faster shortlisting.
- Captured: 2026-09-17

## IDEA-002 - True monthly cost and move-in cash

- Status: `Promoted`
- Task: [014 - Calculate true monthly and move-in costs](tasks/014-true-rental-cost.md)
- Problem: Advertised rent does not show the full monthly cost or the cash
  required before moving in.
- Who benefits: Renters comparing options under a fixed budget.
- Idea: Calculate an estimated monthly range and upfront cash requirement from
  rent, deposit, utilities, parking, service fees, and user-provided usage.
- Inverse check: Never present a precise total when inputs are missing. Label
  every amount as listing-supplied, user-estimated, or unknown.
- Open questions:
  - Which usage assumptions should the user configure?
  - How should mixed per-person, per-unit, and metered charges be displayed?
- Dependency: Reliable arithmetic should follow the Vietnamese rental-price
  normalization proposed in Task 010.
- Expected value: Fewer budget surprises after contacting or viewing a room.
- Captured: 2026-09-17

## IDEA-003 - Viewing pack and decision trail

- Status: `Promoted`
- Task: [015 - Add viewing packs and a decision trail](tasks/015-viewing-decision-trail.md)
- Problem: Important details are discovered during a viewing and then forgotten
  or mixed up with another rental.
- Who benefits: Renters visiting several rooms over multiple days.
- Idea: Generate a viewing checklist from missing fields and saved preferences,
  record answers, track a decision status, and draft follow-up questions that
  remain unresolved.
- Inverse check: Do not label a listing or contact as fraudulent from weak
  signals. Show unverified claims and conflicting evidence instead.
- Open questions:
  - Which physical and contract checks belong in the default checklist?
  - Should the first version store notes only, without photos or documents?
- Expected value: Fewer repeated questions and better decisions after several
  viewings.
- Captured: 2026-09-17

-- V16: Kitty Party Schedule Invitations & Date Voting
-- Tracks per-member RSVP status for each schedule entry and the
-- date-negotiation loop (member proposes → all members vote → host confirms).

-- ── schedule_invitations ─────────────────────────────────────────────────────
-- One row per (schedule_entry, member). Created in bulk when the host
-- sends invitations. Members accept or propose a new date.
CREATE TABLE schedule_invitations (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id   UUID        NOT NULL REFERENCES kitty_schedule(id) ON DELETE CASCADE,
    member_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status        VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                              CHECK (status IN ('PENDING','ACCEPTED','DATE_PROPOSED')),
    responded_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, member_id)
);

-- ── date_proposals ───────────────────────────────────────────────────────────
-- Each time a member proposes an alternative date, a row is inserted here.
-- Any earlier OPEN proposal for the same schedule_entry is SUPERSEDED.
CREATE TABLE date_proposals (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id   UUID        NOT NULL REFERENCES kitty_schedule(id) ON DELETE CASCADE,
    proposed_by   UUID        NOT NULL REFERENCES users(id),
    proposed_date DATE        NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                              CHECK (status IN ('OPEN','ACCEPTED','REJECTED','SUPERSEDED')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── date_proposal_votes ──────────────────────────────────────────────────────
-- One vote per (proposal, member). Members vote ACCEPT or REJECT.
-- When every group member has voted ACCEPT the proposal is auto-accepted.
CREATE TABLE date_proposal_votes (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    proposal_id UUID        NOT NULL REFERENCES date_proposals(id) ON DELETE CASCADE,
    member_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote        VARCHAR(10) NOT NULL CHECK (vote IN ('ACCEPT','REJECT')),
    voted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (proposal_id, member_id)
);

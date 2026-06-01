# VISION — mmo-rpg (future scope, NOT v1)

> This file exists so the big ideas have a home and stay out of the v1 code.
> Nothing here is built in v1. When the v1 vertical slice (GitHub → character →
> Explore) is finished and solid, revisit this list and pick the next ONE thing.
> The point of writing them down is to let the mind release them — not to build them.

The v1 architecture is designed so these can be added without a rewrite. In particular,
the **scoring engine is an isolated module**: new signal sources plug into it. And the
product is split conceptually into a hunter-facing side and a (future) company-facing
side from day one.

---

## The full vision (what this could become)

A two-sided professional-identity platform, framed as an MMORPG, where reputation is
artifact-driven and verification-aware. Below are the pieces, roughly in the order they'd
likely make sense to add — each one is essentially its own product, which is exactly why
none of them are in v1.

### 1. Two-sided marketplace — company/recruiter-facing product
A completely separate product surface for companies who are drowning in bot-generated,
spray-and-pray applications and want **verified, niche, high-signal** candidates. Search
and filter hunters by verified dimensions, view a profile-as-seen-by-a-company,
shortlist, contact. This is the most defensible monetizable direction: the verified
hiring signal is the moat. The whole v1 (verified scoring) is the foundation this sits on.

### 2. Public shareable profiles
Each hunter's character profile at its own public URL, shareable like a LinkedIn profile
or a GitHub README badge. Easy technical extension of v1; deliberately deferred to keep
v1 focused.

### 3. More signal sources (still dev-focused)
Beyond GitHub: blog posts, conference talks, published packages on more registries,
Stack Overflow, etc. Each must be slotted into the **verification hierarchy**:
inspectable + attributable (GitHub-tier) > inspectable but not attributable (a portfolio
link) > self-reported (a CV bullet). The Trust meta-score already models this; new
sources just declare which tier they sit in.

### 4. Multi-profession expansion (designers, marketers, PMs)
The original "universal" vision. The hard part is NOT "find a platform per profession" —
it's that **a platform existing ≠ its artifacts being verifiable**. GitHub is verifiable
because the artifact is *inspectable* (you can read the code) AND the contribution is
*attributable* (commits are bound to identity). Dribbble/Behance are galleries: anyone
can upload anyone's work → not proof. Each new profession needs a real answer to
"what makes a signal here trustworthy?" before it's worth adding. Do not add a profession
until that's solved for it.

### 5. Social hub / engagement layer
A feed where hunters share what they've built, exchange knowledge, and earn points.
**Caution (this is why it's deferred, not just descoped):** a social/engagement layer is
a volume-of-activity machine, and volume is exactly the spam vector that erodes the
verified signal which makes a profile valuable to a recruiter. The thing that drives
retention (posting, points) can corrode the thing that makes the platform trustworthy.
If built, it must be designed so engagement does NOT inflate the hiring signal — keep the
two scores separate.

### 6. Open-source collaboration hub
Hunters post "I'm building X, seeking contributors"; community members contribute and
earn points. **Caution:** double chicken-and-egg problem — you need critical mass of
people before posting is worth it, AND points that buy something somewhere before
contributing is worth the effort. With zero users (where you start), both are dead. Also,
GitHub already does collaboration; you won't pull people there with points that purchase
nothing. This needs a live community AND a points economy with real payoff first.

### 7. Figurative AI character art
Rendered character portraits via Midjourney / DALL·E / Stable Diffusion instead of (or
alongside) the procedural emblems. Bigger raw visual "wow" for a fantasy theme. Deferred
because: it's stochastic (hard to keep a consistent visual language: same build → same
look), it's a separate async pipeline with per-generation cost that scales with users,
storage, latency in onboarding, and output moderation. The v1 emblem route wins on
control, consistency, scale, and zero cost. Add figurative art only if the "epic
portrait" moment proves worth the operational cost.

---

## The recurring trap to remember

Each of the above is a separate company. The temptation is to build them simultaneously
because they form a coherent *vision*. They do not form a buildable *v1*. Finish the
vertical slice first. The vision stays big (it helps in interviews and any future pitch);
the first thing built stays small (so it actually exists).

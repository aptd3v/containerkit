#!/usr/bin/env python3
"""Floor for hand-written pkg/kit statements.

The cover profile still contains pkg/kit/tsx_plan_gen.go. That generated
decoder is not part of the floor. Taskflow coverage is a separate gate.

Usage:
  scripts/ci/kitcover.py --self-test
  scripts/ci/kitcover.py --floor 50 kit.out
"""

import argparse
import sys
from pathlib import Path


GENERATED = "tsx_plan_gen.go"


def statement_totals(text):
    hand_total = hand_covered = 0
    gen_total = gen_covered = 0
    for line in text.splitlines():
        if not line or line.startswith("mode:"):
            continue
        file, rest = line.split(":", 1)
        _span, num, count = rest.rsplit(" ", 2)
        statements = int(num)
        covered = statements if int(count) > 0 else 0
        if file.endswith("/" + GENERATED) or file.endswith("\\" + GENERATED):
            gen_total += statements
            gen_covered += covered
        else:
            hand_total += statements
            hand_covered += covered
    return hand_covered, hand_total, gen_covered, gen_total


def percent(covered, total):
    if total == 0:
        return 0.0
    return 100.0 * covered / total


def report(path, floor):
    text = Path(path).read_text()
    hand_covered, hand_total, gen_covered, gen_total = statement_totals(text)
    hand = percent(hand_covered, hand_total)
    gen = percent(gen_covered, gen_total)
    print(
        f"pkg/kit hand-written statements: {hand:.1f}% ({hand_covered}/{hand_total})"
    )
    print(
        f"pkg/kit generated decoder {GENERATED}: {gen:.1f}% ({gen_covered}/{gen_total})"
    )
    print(f"floor: {floor}% of hand-written statements")
    if hand_total == 0:
        print("cover profile has no hand-written statements", file=sys.stderr)
        return 1
    if hand_covered * 100 < floor * hand_total:
        print(
            f"hand-written coverage {hand:.1f}% is below {floor}%",
            file=sys.stderr,
        )
        return 1
    return 0


def self_test():
    above = "\n".join(
        [
            "mode: set",
            "github.com/aptd3v/containerkit/pkg/kit/foo.go:1.1,2.2 10 1",
            "github.com/aptd3v/containerkit/pkg/kit/tsx_plan_gen.go:1.1,2.2 90 0",
        ]
    )
    below = "\n".join(
        [
            "mode: set",
            "github.com/aptd3v/containerkit/pkg/kit/foo.go:1.1,2.2 4 1",
            "github.com/aptd3v/containerkit/pkg/kit/foo.go:3.1,4.2 6 0",
            "github.com/aptd3v/containerkit/pkg/kit/tsx_plan_gen.go:1.1,2.2 10 1",
        ]
    )
    hand_c, hand_t, gen_c, gen_t = statement_totals(above)
    if (hand_c, hand_t, gen_c, gen_t) != (10, 10, 0, 90):
        print(f"above totals {(hand_c, hand_t, gen_c, gen_t)}", file=sys.stderr)
        return 1
    if hand_c * 100 < 50 * hand_t:
        print("above-floor profile was rejected", file=sys.stderr)
        return 1
    hand_c, hand_t, gen_c, gen_t = statement_totals(below)
    if (hand_c, hand_t, gen_c, gen_t) != (4, 10, 10, 10):
        print(f"below totals {(hand_c, hand_t, gen_c, gen_t)}", file=sys.stderr)
        return 1
    if not (hand_c * 100 < 50 * hand_t):
        print("below-floor profile was accepted", file=sys.stderr)
        return 1
    print("kitcover self-test ok")
    return 0


def main(argv):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--floor", type=int, default=50)
    parser.add_argument("--self-test", action="store_true")
    parser.add_argument("profile", nargs="?")
    args = parser.parse_args(argv)
    if args.self_test:
        return self_test()
    if not args.profile:
        parser.error("profile is required")
    if args.floor < 0 or args.floor > 100:
        parser.error("--floor must be 0..100")
    return report(args.profile, args.floor)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

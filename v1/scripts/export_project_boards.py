#!/usr/bin/env python3

import pathlib
import subprocess
import sys

CONTAINER_NAME = "agentic-repositories-postgres"
DB_USER = "postgres"
DB_NAME = "agentic_orchestrator"
OUTPUT_FILE = pathlib.Path(__file__).resolve().parents[1] / "board_export.json"

SQL = r"""
WITH boards AS (
    SELECT b.id, b.project_id, b.name, b.state, b.created_at, b.updated_at
    FROM project_boards b
    ORDER BY b.updated_at DESC, b.id
)
SELECT jsonb_pretty(
    COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'id', b.id,
                'project_id', b.project_id,
                'name', b.name,
                'state', b.state,
                'created_at', b.created_at,
                'updated_at', b.updated_at,
                'epics', COALESCE((
                    SELECT jsonb_agg(
                        jsonb_build_object(
                            'id', e.id,
                            'board_id', e.board_id,
                            'title', e.title,
                            'objective', e.objective,
                            'state', e.state,
                            'rank', e.rank,
                            'depends_on_epic_ids', e.depends_on_epic_ids,
                            'created_at', e.created_at,
                            'updated_at', e.updated_at,
                            'tasks', COALESCE((
                                SELECT jsonb_agg(to_jsonb(t.*) ORDER BY t.rank, t.created_at, t.id)
                                FROM project_board_tasks t
                                WHERE t.board_id = e.board_id AND t.epic_id = e.id
                            ), '[]'::jsonb)
                        )
                        ORDER BY e.rank, e.created_at, e.id
                    )
                    FROM project_board_epics e
                    WHERE e.board_id = b.id
                ), '[]'::jsonb)
            )
        ),
        '[]'::jsonb
    )
)
FROM boards b;
"""


def main() -> int:
        command = [
                "docker",
                "exec",
                CONTAINER_NAME,
                "psql",
                "-U",
                DB_USER,
                "-d",
                DB_NAME,
                "-At",
                "-c",
                SQL,
        ]

        result = subprocess.run(command, check=False, capture_output=True, text=True)
        if result.returncode != 0:
                message = result.stderr.strip() or result.stdout.strip() or "psql export failed"
                print(f"Export failed: {message}", file=sys.stderr)
                return 1

        OUTPUT_FILE.write_text(result.stdout, encoding="utf-8")
        print(f"Wrote board export to {OUTPUT_FILE}")
        return 0


if __name__ == "__main__":
        raise SystemExit(main())

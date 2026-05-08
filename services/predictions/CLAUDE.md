# services/predictions

Python FastAPI service for predictive budget analytics. Called by the backend API — not exposed directly to the frontend.

## Commands

| Command | Description |
|---------|-------------|
| `python -m venv .venv` | Create virtual environment (first time only) |
| `.venv\Scripts\activate` | Activate virtual environment (Windows) |
| `pip install -r requirements.txt` | Install dependencies |
| `uvicorn main:app --reload --port 8001` | Start dev server (port 8001) |
| `pytest` | Run tests |

## Structure

To be documented once the service is scaffolded.

## Conventions

- Endpoints: defined in `main.py` or `routers/`
- ML models/logic: `services/` or `models/`
- Tests: `test_*.py` files, run with `pytest`
- Dependencies: pinned in `requirements.txt`

## API surface

To be documented as endpoints are added.

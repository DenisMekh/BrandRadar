from dev import (
    set_seed,
    load_config,
    create_or_load_model,
    prepare_data,
    evaluate,
    show_errors,
    test_contrastive,
    test_simple,
)


if __name__ == "__main__":
    cfg = load_config()
    set_seed(cfg["train"]["seed"])

    sm = create_or_load_model(cfg["model"]["name"])

    test_simple(sm)
    test_contrastive(sm)

    print("\nPreparing data...")
    df = prepare_data(cfg, sm)
    acc, y_true, y_pred, all_probs = evaluate(sm, df)

    show_errors(sm, df, y_true, y_pred)

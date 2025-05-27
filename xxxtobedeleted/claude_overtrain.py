

# Use the full dataset for training
train_dataloader = DataLoader(
    full_dataset,
    batch_size=4,
    shuffle=True,
    num_workers=2,
    collate_fn=lambda batch: {
        "input_values": torch.stack([item["input_values"] for item in batch]),
        "labels": torch.stack([item["labels"] for item in batch]),
        "text": [item["text"] for item in batch]
    }
)

# Learning rate scheduler with cosine annealing
from torch.optim.lr_scheduler import CosineAnnealingLR

optimizer = torch.optim.AdamW(model.parameters(), lr=2e-5)
num_epochs = 80  # More epochs for memorization
scheduler = CosineAnnealingLR(optimizer, T_max=num_epochs, eta_min=1e-6)

# Training loop
best_loss = float('inf')
for epoch in range(num_epochs):
    model.train()
    train_loss = 0

    for batch in train_dataloader:
        input_values = batch["input_values"].to(device)
        labels = batch["labels"].to(device)

        outputs = model(input_values=input_values, labels=labels)
        loss = outputs.loss

        optimizer.zero_grad()
        loss.backward()
        optimizer.step()

        train_loss += loss.item()

    avg_train_loss = train_loss / len(train_dataloader)
    scheduler.step()

    print(f"Epoch {epoch+1}/{num_epochs}, Train Loss: {avg_train_loss:.4f}, LR: {scheduler.get_last_lr()[0]:.2e}")

    # Save best model
    if avg_train_loss < best_loss:
        best_loss = avg_train_loss
        model.save_adapter(f"./saved_adapters/best_model/", adapter_name)

    # Periodically check CER/WER on a few examples
    if (epoch + 1) % 10 == 0:
        model.eval()
        with torch.no_grad():
            for i in range(min(5, len(full_dataset))):
                sample = full_dataset[i]
                input_values = sample["input_values"].unsqueeze(0).to(device)

                outputs = model(input_values=input_values)
                predicted_ids = torch.argmax(outputs.logits, dim=-1)

                transcription = processor.batch_decode(predicted_ids)[0]
                reference = sample["text"]

                print(f"Example {i+1}:")
                print(f"  Reference: {reference}")
                print(f"  Predicted: {transcription}")
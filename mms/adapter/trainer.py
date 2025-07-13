import torch
import torch.nn as nn
from torch.utils.data import DataLoader
from transformers import get_linear_schedule_with_warmup
import logging
from tqdm import tqdm

#
# This is sample code provided by Claude 7/12/25
#

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def train_mms_adapter(model, dataset, num_epochs=3, lr=5e-5,
                     warmup_steps=100, max_grad_norm=1.0, log_steps=50):
    """
    Train MMS adapter with raw PyTorch

    Args:
        model: Your MMS model with adapter
        dataset: Your custom dataset that returns padded batches as dicts
        num_epochs: Number of training epochs
        lr: Learning rate
        warmup_steps: Warmup steps for lr scheduler
        max_grad_norm: Max gradient norm for clipping
        log_steps: Steps between logging
    """

    # Setup data loader
    dataloader = DataLoader(dataset, batch_size=batch_size, shuffle=True)

    # Setup optimizer and scheduler
    optimizer = torch.optim.AdamW(model.parameters(), lr=lr)
    total_steps = len(dataloader) * num_epochs
    scheduler = get_linear_schedule_with_warmup(
        optimizer, num_warmup_steps=warmup_steps, num_training_steps=total_steps
    )

    # Training loop
    model.train()
    global_step = 0
    total_loss = 0

    for epoch in range(num_epochs):
        logger.info(f"Starting epoch {epoch + 1}/{num_epochs}")
        epoch_loss = 0

        for step, batch in enumerate(tqdm(dataloader, desc=f"Epoch {epoch + 1}")):
            # Move batch to device (assumes batch is dict with tensor values)
            batch = batch[0] if isinstance(batch, (list, tuple)) else batch
            batch = {k: v.to(model.device) if isinstance(v, torch.Tensor) else v
                    for k, v in batch.items()}


            # Forward pass
            outputs = model(**batch)
            loss = outputs.loss

            # Backward pass
            loss.backward()

            # Gradient clipping
            torch.nn.utils.clip_grad_norm_(model.parameters(), max_grad_norm)

            # Optimizer step
            optimizer.step()
            scheduler.step()
            optimizer.zero_grad()

            # Accumulate loss
            epoch_loss += loss.item()
            total_loss += loss.item()
            global_step += 1

            # Logging
            if global_step % log_steps == 0:
                avg_loss = total_loss / global_step
                current_lr = scheduler.get_last_lr()[0]
                logger.info(
                    f"Step {global_step}: Loss = {loss.item():.4f}, "
                    f"Avg Loss = {avg_loss:.4f}, LR = {current_lr:.2e}"
                )

        # End of epoch logging
        avg_epoch_loss = epoch_loss / len(dataloader)
        logger.info(f"Epoch {epoch + 1} completed. Average loss: {avg_epoch_loss:.4f}")

    logger.info("Training completed!")
    return model

# Usage example:
# trained_model = train_mms_adapter(
#     model=your_mms_model,
#     dataset=your_custom_dataset,
#     num_epochs=3,
#     lr=5e-5
# )
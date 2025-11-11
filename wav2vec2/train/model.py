from transformers import Wav2Vec2Config, Wav2Vec2ForCTC

# This code was provided by Claude.ai.
# It claimed that these settings were the correct
# way to create a raw wav2vec2 model to train from scratch

def getWav2Vec2ForCTCModel(processor):
    # Create the configuration
    config = Wav2Vec2Config(
        # Audio processing
        conv_dim=(512, 512, 512, 512, 512, 512, 512),  # CNN feature extraction layers
        conv_stride=(5, 2, 2, 2, 2, 2, 2),             # Stride for each conv layer
        conv_kernel=(10, 3, 3, 3, 3, 2, 2),            # Kernel size for each conv layer

        # Transformer layers
        hidden_size=768,                               # Size of hidden representations
        num_hidden_layers=12,                          # Number of transformer layers
        num_attention_heads=12,                        # Number of attention heads
        intermediate_size=3072,                        # Size of feed-forward layers

        # Your vocabulary
        vocab_size=len(processor.tokenizer),           # Size of your tokenizer vocabulary

        # CTC specific
        pad_token_id=processor.tokenizer.pad_token_id, # Padding token ID

        # Training parameters (same ones you're already using)
        mask_time_prob=0.01,                           # Your reduced masking
        mask_time_length=1, #originally 2,             # Your shorter masks
        mask_feature_prob=0.0,                         # Your disabled feature masking

        # Additional useful settings
        ctc_loss_reduction="mean",
        apply_spec_augment=True,                       # Enable/disable augmentation
    )
    # Create model from scratch with this config
    model = Wav2Vec2ForCTC(config)
    return model
import type { User } from '../types/auth'
import type { Node } from '../types/graph'

export const mockUser: User = {
  id: '1',
  username: 'admin',
  role: 'admin',
  created_at: '2024-01-01T00:00:00Z',
}

export const mockCredentials = {
  username: 'admin',
  password: 'admin123',
}

export const mockNodes: Node[] = [
  {
    id: 'transformer',
    name: 'Transformer',
    category: 'dl_arch',
    description: 'A neural network architecture based on self-attention mechanisms',
    influence_score: 9.8,
    first_appeared: 2017,
  },
  {
    id: 'bert',
    name: 'BERT',
    category: 'nlp_model',
    description: 'Bidirectional Encoder Representations from Transformers',
    influence_score: 9.6,
    first_appeared: 2018,
  },
  {
    id: 'gpt',
    name: 'GPT',
    category: 'nlp_model',
    description: 'Generative Pre-trained Transformer based autoregressive model',
    influence_score: 9.7,
    first_appeared: 2018,
  },
  {
    id: 'gpt4',
    name: 'GPT-4',
    category: 'nlp_model',
    description: 'OpenAI large-scale multimodal language model',
    influence_score: 9.9,
    first_appeared: 2023,
  },
  {
    id: 'gan',
    name: 'GAN',
    category: 'gen_model',
    description: 'Generative Adversarial Networks',
    influence_score: 8.9,
    first_appeared: 2014,
  },
  {
    id: 'cnn',
    name: 'CNN',
    category: 'dl_arch',
    description: 'Convolutional Neural Networks',
    influence_score: 9.0,
    first_appeared: 1989,
  },
  {
    id: 'resnet',
    name: 'ResNet',
    category: 'cv_model',
    description: 'Residual Networks, solving deep network degradation',
    influence_score: 9.3,
    first_appeared: 2015,
  },
  {
    id: 'stable_diffusion',
    name: 'Stable Diffusion',
    category: 'gen_model',
    description: 'Text-to-image generation model based on diffusion',
    influence_score: 9.4,
    first_appeared: 2022,
  },
  {
    id: 'vit',
    name: 'ViT',
    category: 'cv_model',
    description: 'Vision Transformer for image classification',
    influence_score: 9.1,
    first_appeared: 2020,
  },
  {
    id: 'clip',
    name: 'CLIP',
    category: 'multimodal',
    description: 'Contrastive Language-Image Pre-training model',
    influence_score: 9.0,
    first_appeared: 2021,
  },
]

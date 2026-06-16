import type { GraphNode, GraphEdge, Community, HierarchicalCommunity } from '../types/graph'
import { LEVEL_PALETTES } from '../types/graph'

export const rawNodes: (Omit<GraphNode, 'community_id' | 'degree'> & { first_appeared: string })[] = [
  // ══════════════════════════════════════════════
  // Deep Learning Foundations
  // ══════════════════════════════════════════════
  { id: 'mlp', name: 'MLP', description: 'Multi-layer perceptron, the basic form of neural networks', influence_score: 8.0, first_appeared: '1957-01' },
  { id: 'cnn', name: 'CNN', description: 'Convolutional neural network, the cornerstone of computer vision', influence_score: 9.3, first_appeared: '1989-01' },
  { id: 'rnn', name: 'RNN', description: 'Recurrent neural network, foundational architecture for sequential data', influence_score: 8.0, first_appeared: '1986-01' },
  { id: 'lstm', name: 'LSTM', description: 'Long short-term memory network, solves the long-term dependency problem', influence_score: 8.5, first_appeared: '1997-01' },
  { id: 'autoencoder', name: 'Autoencoder', description: 'Autoencoder, a classic architecture for unsupervised representation learning', influence_score: 7.8, first_appeared: '1986-01' },
  { id: 'gru', name: 'GRU', description: 'Gated recurrent unit, a simplified and efficient variant of LSTM', influence_score: 7.5, first_appeared: '2014-01' },
  { id: 'encoder_decoder', name: 'Encoder-Decoder', description: 'Encoder-decoder architecture, the foundational paradigm for sequence transformation', influence_score: 8.4, first_appeared: '2014-01' },
  { id: 'seq2seq', name: 'Seq2Seq', description: 'Sequence-to-sequence model, core of machine translation and similar tasks', influence_score: 8.0, first_appeared: '2014-01' },

  // ══════════════════════════════════════════════
  // Transformer Ecosystem
  // ══════════════════════════════════════════════
  { id: 'transformer', name: 'Transformer', description: 'Self-attention-based architecture that reshaped the entire AI landscape, paper cited over 140k times', influence_score: 9.9, first_appeared: '2017-01' },
  { id: 'attention', name: 'Attention', description: 'Attention mechanism, enabling models to selectively focus', influence_score: 9.5, first_appeared: '2014-01' },
  { id: 'self_attention', name: 'Self-Attention', description: 'Self-attention, attention computation among elements within a sequence', influence_score: 9.3, first_appeared: '2017-01' },
  { id: 'multi_head_attention', name: 'Multi-Head Attention', description: 'Multi-head attention, capturing features in different subspaces in parallel', influence_score: 9.0, first_appeared: '2017-01' },
  { id: 'attention_mechanism', name: 'Attention Mechanism', description: 'Theoretical framework of attention, selectively focusing on input', influence_score: 9.2, first_appeared: '2014-01' },
  { id: 'positional_encoding', name: 'Positional Encoding', description: 'Positional encoding, injecting sequence position information into Transformer', influence_score: 8.4, first_appeared: '2017-01' },
  { id: 'rope', name: 'RoPE', description: 'Rotary position encoding, adopted by mainstream models such as LLaMA', influence_score: 8.2, first_appeared: '2021-01' },
  { id: 'gqa', name: 'Grouped Query Attention', description: 'Grouped query attention, reducing KV cache memory usage', influence_score: 7.8, first_appeared: '2023-01' },
  { id: 'flash_attention', name: 'Flash Attention', description: 'Efficient attention computation, reducing memory access and accelerating large model training and inference', influence_score: 8.6, first_appeared: '2022-01' },
  { id: 'kv_cache', name: 'KV Cache', description: 'Key-value cache, accelerating autoregressive inference', influence_score: 7.9, first_appeared: '2018-01' },
  { id: 'speculative_decoding', name: 'Speculative Decoding', description: 'Speculative decoding, using a small model to accelerate large model inference', influence_score: 7.5, first_appeared: '2023-01' },

  // ══════════════════════════════════════════════
  // Classic NLP Models
  // ══════════════════════════════════════════════
  { id: 'bert', name: 'BERT', description: 'Bidirectional encoder representations from transformers, initiating the pretrain-finetune paradigm', influence_score: 9.5, first_appeared: '2018-01' },
  { id: 'roberta', name: 'RoBERTa', description: 'Optimized version of BERT with more robust pre-training methods', influence_score: 8.5, first_appeared: '2019-01' },
  { id: 't5', name: 'T5', description: 'Text-to-Text Transfer Transformer unified framework', influence_score: 8.8, first_appeared: '2019-01' },

  // ══════════════════════════════════════════════
  // Large Language Models
  // ══════════════════════════════════════════════
  { id: 'llm', name: 'LLM', description: 'Large language model, large-scale pre-trained language model', influence_score: 9.6, first_appeared: '2020-01' },
  { id: 'gpt', name: 'GPT', description: 'Generative pre-trained Transformer, ushering in the era of large models', influence_score: 9.4, first_appeared: '2018-01' },
  { id: 'gpt2', name: 'GPT-2', description: '1.5B parameter generative model, initially withheld from open release due to concerns', influence_score: 8.7, first_appeared: '2019-01' },
  { id: 'gpt3', name: 'GPT-3', description: '175B parameter large model, demonstrating emergent abilities', influence_score: 9.3, first_appeared: '2020-01' },
  { id: 'gpt4', name: 'GPT-4', description: 'Multimodal large language model, approaching human-level performance', influence_score: 9.7, first_appeared: '2023-01' },
  { id: 'chatgpt', name: 'ChatGPT', description: 'OpenAI conversational AI assistant, reached 100M users in two months', influence_score: 9.8, first_appeared: '2022-01' },
  { id: 'llama', name: 'LLaMA', description: 'Meta open-source large language model, sparking the open-source LLM wave', influence_score: 9.0, first_appeared: '2023-01' },
  { id: 'gemini', name: 'Gemini', description: 'Google multimodal large model', influence_score: 9.0, first_appeared: '2023-01' },
  { id: 'claude', name: 'Claude', description: 'Anthropic AI assistant, known for safety alignment', influence_score: 8.8, first_appeared: '2023-01' },
  { id: 'mistral', name: 'Mistral', description: 'Mistral AI open-source large model', influence_score: 8.5, first_appeared: '2023-01' },
  { id: 'deepseek', name: 'DeepSeek', description: 'DeepSeek large model, efficient open-source model', influence_score: 8.6, first_appeared: '2024-01' },
  { id: 'qwen', name: 'Qwen', description: 'Alibaba Tongyi Qianwen large model series', influence_score: 8.3, first_appeared: '2023-01' },
  { id: 'grok', name: 'Grok', description: 'xAI large language model', influence_score: 7.8, first_appeared: '2023-01' },
  { id: 'phi', name: 'Phi', description: 'Microsoft small-parameter high-performance language model', influence_score: 8.0, first_appeared: '2023-01' },
  { id: 'yi', name: 'Yi', description: '01.AI large model', influence_score: 7.8, first_appeared: '2023-01' },

  // ══════════════════════════════════════════════
  // NLP Techniques
  // ══════════════════════════════════════════════
  { id: 'word2vec', name: 'Word2Vec', description: 'Word vector representation learning, driving the NLP representation revolution', influence_score: 8.6, first_appeared: '2013-01' },
  { id: 'glove', name: 'GloVe', description: 'Global vectors for word representation', influence_score: 7.8, first_appeared: '2014-01' },
  { id: 'fasttext', name: 'fastText', description: 'Facebook library for text classification and subword representation', influence_score: 7.4, first_appeared: '2016-01' },
  { id: 'tokenization', name: 'Tokenization', description: 'Text tokenization, splitting text into model input units', influence_score: 7.6, first_appeared: '1990-01' },
  { id: 'bpe', name: 'BPE', description: 'Byte-pair encoding, a subword tokenization algorithm', influence_score: 8.0, first_appeared: '2016-01' },
  { id: 'embedding', name: 'Embedding', description: 'Mapping discrete symbols to continuous vector space', influence_score: 8.7, first_appeared: '2013-01' },
  { id: 'prompt_engineering', name: 'Prompt Engineering', description: 'Prompt engineering, optimizing model input to improve output', influence_score: 8.8, first_appeared: '2020-01' },
  { id: 'chain_of_thought', name: 'Chain-of-Thought', description: 'Chain-of-thought prompting, guiding models to reason step by step', influence_score: 9.0, first_appeared: '2022-01' },
  { id: 'rag', name: 'RAG', description: 'Retrieval-augmented generation, enabling LLMs to access external knowledge', influence_score: 9.0, first_appeared: '2023-01' },
  { id: 'in_context_learning', name: 'In-Context Learning', description: 'In-context learning, adaptation without parameter updates', influence_score: 8.9, first_appeared: '2020-01' },
  { id: 'function_calling', name: 'Function Calling', description: 'The ability of LLMs to call external tools/APIs', influence_score: 8.4, first_appeared: '2023-01' },
  { id: 'tool_use', name: 'Tool Use', description: 'The paradigm of AI using external tools to enhance capabilities', influence_score: 8.2, first_appeared: '2023-01' },
  { id: 'instruction_tuning', name: 'Instruction Tuning', description: 'Instruction tuning, teaching models to follow human instructions', influence_score: 8.6, first_appeared: '2021-01' },
  { id: 'beam_search', name: 'Beam Search', description: 'Beam search decoding strategy, generating better sequences', influence_score: 7.5, first_appeared: '2014-01' },
  { id: 'temperature', name: 'Temperature', description: 'Temperature parameter, controlling randomness and diversity of generated output', influence_score: 7.7, first_appeared: '2018-01' },
  { id: 'top_k', name: 'Top-K Sampling', description: 'Top-K sampling strategy, limiting the candidate token range', influence_score: 7.4, first_appeared: '2018-01' },
  { id: 'top_p', name: 'Top-P Sampling', description: 'Nucleus sampling strategy, dynamically adjusting the candidate token set', influence_score: 7.5, first_appeared: '2019-01' },
  { id: 'hallucination', name: 'Hallucination', description: 'The phenomenon and study of LLMs generating inaccurate or fabricated content', influence_score: 8.4, first_appeared: '2020-01' },
  { id: 'softmax', name: 'Softmax', description: 'Activation function that converts outputs into a probability distribution', influence_score: 8.2, first_appeared: '1989-01' },
  { id: 'cross_entropy', name: 'Cross Entropy', description: 'Cross-entropy loss function, the standard loss for classification tasks', influence_score: 8.0, first_appeared: '1948-01' },

  // ══════════════════════════════════════════════
  // Computer Vision
  // ══════════════════════════════════════════════
  { id: 'resnet', name: 'ResNet', description: 'Residual network, solving the degradation problem in deep networks', influence_score: 9.3, first_appeared: '2015-01' },
  { id: 'vit', name: 'ViT', description: 'Vision Transformer, applying Transformer to images', influence_score: 9.1, first_appeared: '2020-01' },
  { id: 'yolo', name: 'YOLO', description: 'Real-time object detection model series', influence_score: 8.7, first_appeared: '2015-01' },
  { id: 'faster_rcnn', name: 'Faster R-CNN', description: 'Fast object detection based on region proposal networks', influence_score: 8.3, first_appeared: '2015-01' },
  { id: 'mask_rcnn', name: 'Mask R-CNN', description: 'Instance segmentation model, handling both detection and segmentation', influence_score: 8.1, first_appeared: '2017-01' },
  { id: 'unet', name: 'U-Net', description: 'Encoder-decoder semantic segmentation network', influence_score: 8.5, first_appeared: '2015-01' },
  { id: 'sam', name: 'SAM', description: 'Segment Anything Model, universal image segmentation', influence_score: 8.8, first_appeared: '2023-01' },
  { id: 'detr', name: 'DETR', description: 'End-to-end object detection model based on Transformer', influence_score: 8.2, first_appeared: '2020-01' },
  { id: 'neRF', name: 'NeRF', description: 'Neural radiance field, 3D scene reconstruction', influence_score: 8.4, first_appeared: '2020-01' },
  { id: 'object_detection', name: 'Object Detection', description: 'Object detection task, identifying object categories and positions in images', influence_score: 8.2, first_appeared: '2012-01' },
  { id: 'image_segmentation', name: 'Image Segmentation', description: 'Image segmentation, pixel-level classification', influence_score: 8.0, first_appeared: '2012-01' },
  { id: 'image_classification', name: 'Image Classification', description: 'Image classification, determining the category of an image', influence_score: 8.0, first_appeared: '2012-01' },
  { id: 'ocr', name: 'OCR', description: 'Optical character recognition', influence_score: 7.5, first_appeared: '1990-01' },
  { id: 'patch_embedding', name: 'Patch Embedding', description: 'Splitting an image into fixed-size patches and embedding them as vectors', influence_score: 7.8, first_appeared: '2020-01' },
  { id: 'super_resolution', name: 'Super Resolution', description: 'Image super-resolution reconstruction', influence_score: 7.7, first_appeared: '2014-01' },
  { id: 'style_transfer', name: 'Style Transfer', description: 'Style transfer, applying artistic styles to images', influence_score: 7.5, first_appeared: '2015-01' },

  // ══════════════════════════════════════════════
  // Generative Models
  // ══════════════════════════════════════════════
  { id: 'gan', name: 'GAN', description: 'Generative adversarial network, pioneering a new paradigm for generative AI', influence_score: 9.0, first_appeared: '2014-01' },
  { id: 'vae', name: 'VAE', description: 'Variational autoencoder, probabilistic generative modeling', influence_score: 8.1, first_appeared: '2013-01' },
  { id: 'diffusion_model', name: 'Diffusion Model', description: 'Diffusion generative model, generating high-quality data through iterative denoising', influence_score: 9.2, first_appeared: '2015-01' },
  { id: 'ddpm', name: 'DDPM', description: 'Denoising diffusion probabilistic model, the foundational algorithm of diffusion models', influence_score: 8.7, first_appeared: '2020-01' },
  { id: 'stable_diffusion', name: 'Stable Diffusion', description: 'Open-source image generation model, democratizing AI creation', influence_score: 9.4, first_appeared: '2022-01' },
  { id: 'stylegan', name: 'StyleGAN', description: 'Style-based generative adversarial network, high-quality image generation', influence_score: 8.6, first_appeared: '2018-01' },
  { id: 'midjourney', name: 'Midjourney', description: 'AI image generation tool, leading in aesthetic quality', influence_score: 8.5, first_appeared: '2022-01' },
  { id: 'sora', name: 'Sora', description: 'OpenAI text-to-video generation model', influence_score: 9.0, first_appeared: '2024-01' },
  { id: 'dalle', name: 'DALL-E', description: 'Text-to-image generation model', influence_score: 9.2, first_appeared: '2021-01' },
  { id: 'text_to_image', name: 'Text-to-Image', description: 'Text-to-image generation technology', influence_score: 8.8, first_appeared: '2021-01' },
  { id: 'image_inpainting', name: 'Image Inpainting', description: 'Image inpainting, filling in missing regions', influence_score: 7.4, first_appeared: '2016-01' },
  { id: 'flow_matching', name: 'Flow Matching', description: 'Flow matching generative model, an alternative to diffusion models', influence_score: 7.6, first_appeared: '2023-01' },
  { id: 'score_matching', name: 'Score Matching', description: 'Score matching, the theoretical foundation of diffusion models', influence_score: 7.5, first_appeared: '2015-01' },

  // ══════════════════════════════════════════════
  // Multimodal
  // ══════════════════════════════════════════════
  { id: 'multimodal', name: 'Multimodal', description: 'Multimodal AI, processing text, image, audio and other inputs', influence_score: 9.1, first_appeared: '2020-01' },
  { id: 'clip', name: 'CLIP', description: 'Contrastive language-image pre-training model', influence_score: 9.1, first_appeared: '2021-01' },
  { id: 'blip', name: 'BLIP', description: 'Bootstrapping language-image pre-training model', influence_score: 8.4, first_appeared: '2022-01' },
  { id: 'flamingo', name: 'Flamingo', description: 'DeepMind vision-language model', influence_score: 7.8, first_appeared: '2022-01' },
  { id: 'vision_language', name: 'Vision-Language', description: 'Vision-language model, a multimodal subfield', influence_score: 8.6, first_appeared: '2021-01' },

  // ══════════════════════════════════════════════
  // Speech Processing
  // ══════════════════════════════════════════════
  { id: 'whisper', name: 'Whisper', description: 'OpenAI general-purpose speech recognition model', influence_score: 8.3, first_appeared: '2022-01' },
  { id: 'wav2vec', name: 'wav2vec 2.0', description: 'Self-supervised speech representation learning framework', influence_score: 8.0, first_appeared: '2020-01' },
  { id: 'tts', name: 'TTS', description: 'Text-to-speech synthesis', influence_score: 7.8, first_appeared: '2016-01' },
  { id: 'asr', name: 'ASR', description: 'Automatic speech recognition', influence_score: 7.9, first_appeared: '2006-01' },
  { id: 'music_generation', name: 'Music Generation', description: 'AI music generation', influence_score: 7.6, first_appeared: '2016-01' },

  // ══════════════════════════════════════════════
  // Reinforcement Learning
  // ══════════════════════════════════════════════
  { id: 'reinforcement_learning', name: 'Reinforcement Learning', description: 'Reinforcement learning, learning optimal policies through reward signals', influence_score: 8.8, first_appeared: '1989-01' },
  { id: 'dqn', name: 'DQN', description: 'Deep Q-network, introducing deep learning into reinforcement learning', influence_score: 8.0, first_appeared: '2013-01' },
  { id: 'ppo', name: 'PPO', description: 'Proximal policy optimization, a stable and efficient policy gradient method', influence_score: 8.3, first_appeared: '2017-01' },
  { id: 'alphago', name: 'AlphaGo', description: 'Go AI, defeating the human world champion', influence_score: 9.0, first_appeared: '2016-01' },
  { id: 'q_learning', name: 'Q-Learning', description: 'Classic reinforcement learning algorithm based on value functions', influence_score: 7.8, first_appeared: '1989-01' },
  { id: 'policy_gradient', name: 'Policy Gradient', description: 'Policy gradient method, directly optimizing policies', influence_score: 7.6, first_appeared: '1992-01' },
  { id: 'world_model', name: 'World Model', description: 'World model, learning and simulating environment dynamics', influence_score: 7.6, first_appeared: '2018-01' },
  { id: 'imitation_learning', name: 'Imitation Learning', description: 'Imitation learning, learning from expert demonstrations', influence_score: 7.4, first_appeared: '2009-01' },
  { id: 'sac', name: 'SAC', description: 'Soft actor-critic algorithm, maximizing policy entropy', influence_score: 7.5, first_appeared: '2018-01' },
  { id: 'planning', name: 'Planning', description: 'AI planning capability, decomposing and solving complex tasks', influence_score: 7.8, first_appeared: '1971-01' },

  // ══════════════════════════════════════════════
  // Learning Paradigms
  // ══════════════════════════════════════════════
  { id: 'supervised_learning', name: 'Supervised Learning', description: 'Supervised learning, learning from labeled data', influence_score: 8.5, first_appeared: '1959-01' },
  { id: 'unsupervised_learning', name: 'Unsupervised Learning', description: 'Unsupervised learning, learning from unlabeled data', influence_score: 8.0, first_appeared: '1980-01' },
  { id: 'self_supervised', name: 'Self-Supervised', description: 'Self-supervised learning, constructing supervision signals from the data itself', influence_score: 8.7, first_appeared: '2019-01' },
  { id: 'few_shot', name: 'Few-Shot', description: 'Few-shot learning, adapting with only a few examples', influence_score: 8.1, first_appeared: '2019-01' },
  { id: 'zero_shot', name: 'Zero-Shot', description: 'Zero-shot learning, recognizing without examples', influence_score: 7.9, first_appeared: '2018-01' },
  { id: 'contrastive_learning', name: 'Contrastive Learning', description: 'Contrastive learning, an important paradigm for self-supervised representation learning', influence_score: 8.6, first_appeared: '2020-01' },
  { id: 'curriculum_learning', name: 'Curriculum Learning', description: 'Curriculum learning, training progressively from easy to hard', influence_score: 7.3, first_appeared: '2009-01' },
  { id: 'meta_learning', name: 'Meta-Learning', description: 'Meta-learning, learning how to learn', influence_score: 7.6, first_appeared: '2017-01' },

  // ══════════════════════════════════════════════
  // Training & Optimization
  // ══════════════════════════════════════════════
  { id: 'backpropagation', name: 'Backpropagation', description: 'Backpropagation algorithm, the core of training all deep neural networks', influence_score: 9.8, first_appeared: '1986-01' },
  { id: 'adam', name: 'Adam', description: 'Adaptive optimizer, the most widely used deep learning optimizer', influence_score: 9.0, first_appeared: '2014-01' },
  { id: 'sgd', name: 'SGD', description: 'Stochastic gradient descent, a classic optimization method', influence_score: 8.2, first_appeared: '1951-01' },
  { id: 'gradient_descent', name: 'Gradient Descent', description: 'Gradient descent optimization method, the foundation of all optimization algorithms', influence_score: 8.8, first_appeared: '1951-01' },
  { id: 'learning_rate', name: 'Learning Rate', description: 'Learning rate, the key hyperparameter controlling parameter update step size', influence_score: 7.8, first_appeared: '1951-01' },
  { id: 'dropout', name: 'Dropout', description: 'Dropout regularization, a simple and effective technique to prevent overfitting', influence_score: 8.5, first_appeared: '2012-01' },
];

// Default community id for the mock data. The raw nodes are bucketed
// into a single community; the real portal uses Louvain over the
// returned graph structure to derive meaningful communities.
const DEFAULT_COMMUNITY_ID = 13

export function generateMockData(): {
  nodes: GraphNode[]
  edges: GraphEdge[]
  communities: Community[]
  hierarchicalCommunities: HierarchicalCommunity[]
  allCommunityIds: Map<number, number[]>
} {
  const nodes: GraphNode[] = rawNodes.map((n) => ({
    ...n,
    community_id: DEFAULT_COMMUNITY_ID,
    degree: 3 + Math.floor(Math.random() * 8),
  }))

  const edges: GraphEdge[] = generateEdges(nodes)
  const communities: Community[] = getCommunityList(nodes)
  const { root, allIds } = buildHierarchicalCommunities(nodes)

  return { nodes, edges, communities, hierarchicalCommunities: [root], allCommunityIds: allIds }
}

function generateEdges(nodes: GraphNode[]): GraphEdge[] {
  const edgeSet = new Set<string>()
  const edges: GraphEdge[] = []

  const addEdge = (source: string, target: string, weight: number, relation: string) => {
    if (source === target) return
    const key = [source, target].sort().join('---')
    if (edgeSet.has(key)) return
    edgeSet.add(key)
    edges.push({ source, target, weight: +weight.toFixed(2), relation })
  }

  const byId = new Map(nodes.map(n => [n.id, n]))
  const byCommunity = new Map<number, GraphNode[]>()
  for (const n of nodes) {
    const list = byCommunity.get(n.community_id) || []
    list.push(n)
    byCommunity.set(n.community_id, list)
  }

  const crossBridges: { source: string; target: string; relation: string }[] = [
    { source: 'transformer', target: 'bert', relation: 'architecture foundation' },
    { source: 'transformer', target: 'gpt', relation: 'architecture foundation' },
    { source: 'transformer', target: 'vit', relation: 'applied to CV' },
    { source: 'transformer', target: 'attention', relation: 'core mechanism' },
    { source: 'transformer', target: 'self_attention', relation: 'core mechanism' },
    { source: 'transformer', target: 'multi_head_attention', relation: 'core mechanism' },
    { source: 'transformer', target: 'encoder_decoder', relation: 'improvement' },
    { source: 'transformer', target: 'seq2seq', relation: 'replaces' },
    { source: 'transformer', target: 'llm', relation: 'architecture foundation' },
    { source: 'transformer', target: 'gpt3', relation: 'architecture foundation' },
    { source: 'transformer', target: 'gpt4', relation: 'architecture foundation' },
    { source: 'transformer', target: 'llama', relation: 'architecture foundation' },
    { source: 'transformer', target: 'whisper', relation: 'architecture foundation' },
    { source: 'transformer', target: 'detr', relation: 'applied to CV' },

    { source: 'attention', target: 'bert', relation: 'component' },
    { source: 'attention', target: 'seq2seq', relation: 'enhances' },
    { source: 'attention', target: 'attention_mechanism', relation: 'implementation' },

    { source: 'bert', target: 'tokenization', relation: 'depends on' },
    { source: 'bert', target: 'embedding', relation: 'depends on' },
    { source: 'bert', target: 'encoder_decoder', relation: 'encoder only' },
    { source: 'bert', target: 'roberta', relation: 'optimized variant' },
    { source: 'gpt', target: 'transformer', relation: 'decoder only' },
    { source: 'gpt', target: 'embedding', relation: 'depends on' },
    { source: 'gpt2', target: 'gpt', relation: 'iteration' },
    { source: 'gpt3', target: 'gpt2', relation: 'iteration' },
    { source: 'gpt3', target: 'in_context_learning', relation: 'enables' },
    { source: 'gpt4', target: 'gpt3', relation: 'iteration' },
    { source: 'gpt4', target: 'multimodal', relation: 'implements' },
    { source: 'chatgpt', target: 'gpt3', relation: 'based on' },
    { source: 'chatgpt', target: 'gpt4', relation: 'based on' },
    { source: 'chatgpt', target: 'instruction_tuning', relation: 'depends on' },
    { source: 'chatgpt', target: 'prompt_engineering', relation: 'depends on' },
    { source: 'llm', target: 'chain_of_thought', relation: 'reasoning enhancement' },
    { source: 'llm', target: 'rag', relation: 'knowledge enhancement' },
    { source: 'llm', target: 'hallucination', relation: 'challenge' },
    { source: 'llama', target: 'rope', relation: 'uses' },
    { source: 'llama', target: 'gqa', relation: 'uses' },
    { source: 'gemini', target: 'multimodal', relation: 'implements' },
    { source: 'qwen', target: 'multimodal', relation: 'extends' },

    { source: 'word2vec', target: 'embedding', relation: 'implements' },
    { source: 'glove', target: 'embedding', relation: 'implements' },
    { source: 'chain_of_thought', target: 'gpt3', relation: 'improves reasoning' },
    { source: 'rag', target: 'embedding', relation: 'retrieval dependency' },
    { source: 'rag', target: 'llm', relation: 'enhances' },
    { source: 'instruction_tuning', target: 'supervised_learning', relation: 'based on' },
    { source: 'function_calling', target: 'tool_use', relation: 'implementation' },
    { source: 'bpe', target: 'tokenization', relation: 'implements' },
    { source: 'kv_cache', target: 'gpt', relation: 'accelerates inference' },
    { source: 'flash_attention', target: 'self_attention', relation: 'optimizes' },
    { source: 'flash_attention', target: 'gqa', relation: 'composable with' },
    { source: 'speculative_decoding', target: 'llm', relation: 'accelerates inference' },

    { source: 'cnn', target: 'resnet', relation: 'base architecture' },
    { source: 'cnn', target: 'yolo', relation: 'base architecture' },
    { source: 'cnn', target: 'faster_rcnn', relation: 'base architecture' },
    { source: 'resnet', target: 'vit', relation: 'surpassed by' },
    { source: 'vit', target: 'patch_embedding', relation: 'depends on' },
    { source: 'vit', target: 'transformer', relation: 'architecture foundation' },
    { source: 'vit', target: 'self_attention', relation: 'uses' },
    { source: 'vit', target: 'image_classification', relation: 'used for' },
    { source: 'detr', target: 'transformer', relation: 'architecture foundation' },
    { source: 'detr', target: 'object_detection', relation: 'used for' },
    { source: 'unet', target: 'cnn', relation: 'based on' },
    { source: 'unet', target: 'stable_diffusion', relation: 'backbone network' },
    { source: 'sam', target: 'vit', relation: 'image encoder' },
    { source: 'sam', target: 'image_segmentation', relation: 'used for' },
    { source: 'mask_rcnn', target: 'faster_rcnn', relation: 'extends' },
    { source: 'mask_rcnn', target: 'image_segmentation', relation: 'used for' },
    { source: 'yolo', target: 'object_detection', relation: 'used for' },
    { source: 'neRF', target: 'mlp', relation: 'uses' },
    { source: 'style_transfer', target: 'cnn', relation: 'based on' },
    { source: 'super_resolution', target: 'cnn', relation: 'based on' },
    { source: 'ocr', target: 'cnn', relation: 'based on' },

    { source: 'gan', target: 'cnn', relation: 'uses CNN' },
    { source: 'stylegan', target: 'gan', relation: 'improves' },
    { source: 'diffusion_model', target: 'score_matching', relation: 'theoretical basis' },
    { source: 'ddpm', target: 'diffusion_model', relation: 'implements' },
    { source: 'stable_diffusion', target: 'ddpm', relation: 'based on' },
    { source: 'stable_diffusion', target: 'clip', relation: 'text encoding' },
    { source: 'stable_diffusion', target: 'text_to_image', relation: 'implements' },
    { source: 'sora', target: 'diffusion_model', relation: 'based on' },
    { source: 'sora', target: 'transformer', relation: 'combined with' },
    { source: 'dalle', target: 'clip', relation: 'text encoding' },
    { source: 'dalle', target: 'gpt3', relation: 'combined with' },
    { source: 'dalle', target: 'text_to_image', relation: 'implements' },
    { source: 'midjourney', target: 'diffusion_model', relation: 'based on' },
    { source: 'midjourney', target: 'text_to_image', relation: 'implements' },
    { source: 'flow_matching', target: 'diffusion_model', relation: 'alternative to' },
    { source: 'vae', target: 'autoencoder', relation: 'variant' },
    { source: 'vae', target: 'stable_diffusion', relation: 'latent space' },
    { source: 'image_inpainting', target: 'gan', relation: 'based on' },
    { source: 'image_inpainting', target: 'diffusion_model', relation: 'based on' },

    { source: 'clip', target: 'contrastive_learning', relation: 'based on' },
    { source: 'clip', target: 'vit', relation: 'image encoder' },
    { source: 'clip', target: 'transformer', relation: 'text encoder' },
    { source: 'clip', target: 'multimodal', relation: 'foundation model' },
    { source: 'vision_language', target: 'multimodal', relation: 'subfield' },
    { source: 'multimodal', target: 'vit', relation: 'vision module' },
    { source: 'multimodal', target: 'transformer', relation: 'fusion module' },
    { source: 'blip', target: 'clip', relation: 'improves' },
    { source: 'flamingo', target: 'multimodal', relation: 'implements' },

    { source: 'dqn', target: 'gradient_descent', relation: 'optimization' },
    { source: 'dqn', target: 'backpropagation', relation: 'training' },
    { source: 'ppo', target: 'policy_gradient', relation: 'improves' },
    { source: 'alphago', target: 'reinforcement_learning', relation: 'application' },
    { source: 'alphago', target: 'dqn', relation: 'combined with' },
    { source: 'reinforcement_learning', target: 'q_learning', relation: 'includes' },
    { source: 'reinforcement_learning', target: 'policy_gradient', relation: 'includes' },
    { source: 'sac', target: 'policy_gradient', relation: 'improves' },
    { source: 'imitation_learning', target: 'reinforcement_learning', relation: 'paradigm' },
    { source: 'world_model', target: 'reinforcement_learning', relation: 'extends' },

    { source: 'whisper', target: 'transformer', relation: 'architecture foundation' },
    { source: 'whisper', target: 'asr', relation: 'implements' },
    { source: 'wav2vec', target: 'self_supervised', relation: 'based on' },
    { source: 'wav2vec', target: 'transformer', relation: 'uses' },
    { source: 'tts', target: 'transformer', relation: 'based on' },
    { source: 'music_generation', target: 'transformer', relation: 'based on' },

    { source: 'backpropagation', target: 'mlp', relation: 'training' },
    { source: 'backpropagation', target: 'cnn', relation: 'training' },
    { source: 'backpropagation', target: 'lstm', relation: 'training' },
    { source: 'adam', target: 'transformer', relation: 'training' },
    { source: 'adam', target: 'bert', relation: 'training' },
    { source: 'adam', target: 'gpt', relation: 'training' },
    { source: 'adam', target: 'resnet', relation: 'training' },
    { source: 'adam', target: 'stable_diffusion', relation: 'training' },
    { source: 'adam', target: 'clip', relation: 'training' },
    { source: 'dropout', target: 'transformer', relation: 'regularization' },
    { source: 'dropout', target: 'bert', relation: 'regularization' },
    { source: 'dropout', target: 'cnn', relation: 'regularization' },
    { source: 'gradient_descent', target: 'sgd', relation: 'foundation' },
    { source: 'gradient_descent', target: 'backpropagation', relation: 'combined with' },
    { source: 'sgd', target: 'cnn', relation: 'optimization' },
    { source: 'learning_rate', target: 'adam', relation: 'hyperparameter' },
    { source: 'learning_rate', target: 'sgd', relation: 'hyperparameter' },

    { source: 'mlp', target: 'backpropagation', relation: 'depends on' },
    { source: 'lstm', target: 'rnn', relation: 'improves' },
    { source: 'lstm', target: 'encoder_decoder', relation: 'component' },
    { source: 'gru', target: 'lstm', relation: 'simplified variant' },
    { source: 'encoder_decoder', target: 'seq2seq', relation: 'base architecture' },
    { source: 'seq2seq', target: 'attention', relation: 'enhanced by' },
    { source: 'rnn', target: 'lstm', relation: 'variant base' },
    { source: 'autoencoder', target: 'vae', relation: 'variant' },

    { source: 'embedding', target: 'tokenization', relation: 'next step' },
    { source: 'embedding', target: 'softmax', relation: 'output layer' },
    { source: 'cross_entropy', target: 'softmax', relation: 'paired with' },
    { source: 'self_supervised', target: 'bert', relation: 'training method' },
    { source: 'self_supervised', target: 'contrastive_learning', relation: 'includes' },
    { source: 'few_shot', target: 'gpt3', relation: 'evaluation' },
    { source: 'few_shot', target: 'llm', relation: 'capability' },
    { source: 'zero_shot', target: 'clip', relation: 'capability' },
    { source: 'contrastive_learning', target: 'self_supervised', relation: 'implements' },
    { source: 'contrastive_learning', target: 'simclr', relation: 'method' },
    { source: 'meta_learning', target: 'few_shot', relation: 'method' },
    { source: 'softmax', target: 'temperature', relation: 'parameter' },
    { source: 'top_k', target: 'temperature', relation: 'sampling' },
    { source: 'top_p', target: 'temperature', relation: 'sampling' },
    { source: 'beam_search', target: 'seq2seq', relation: 'decoding' },
    { source: 't5', target: 'encoder_decoder', relation: 'architecture' },
    { source: 't5', target: 'transformer', relation: 'unified framework' },
  ]

  // 1. Add explicit cross-community bridges
  for (const bridge of crossBridges) {
    const source = byId.get(bridge.source)
    const target = byId.get(bridge.target)
    if (source && target) {
      const weight = 0.5 + (source.influence_score + target.influence_score) / 20
      addEdge(bridge.source, bridge.target, weight, bridge.relation)
    }
  }

  // 2. Intra-community random connections (no ring pattern)
  for (const [, communityNodes] of byCommunity) {
    if (communityNodes.length < 2) continue
    const shuffled = [...communityNodes].sort(() => Math.random() - 0.5)
    for (let i = 0; i < shuffled.length; i++) {
      const n = shuffled[i]
      const degreeTarget = Math.min(3, communityNodes.length - 1) + Math.round(n.influence_score - 7)
      const maxEdges = Math.max(2, Math.min(degreeTarget, communityNodes.length - 1))

      for (let j = i + 1; j < shuffled.length && j <= i + maxEdges; j++) {
        const weight = 0.4 + (n.influence_score + shuffled[j].influence_score) / 25
        addEdge(n.id, shuffled[j].id, weight, 'same domain')
      }
    }
  }

  // 3. Random cross-community edges for natural bridging
  const communityIds = [...byCommunity.keys()]
  for (let i = 0; i < communityIds.length; i++) {
    for (let j = i + 1; j < communityIds.length; j++) {
      const commA = byCommunity.get(communityIds[i])!
      const commB = byCommunity.get(communityIds[j])!
      const pairCount = Math.min(2, Math.min(commA.length, commB.length))
      for (let k = 0; k < pairCount; k++) {
        const a = commA[Math.floor(Math.random() * commA.length)]
        const b = commB[Math.floor(Math.random() * commB.length)]
        addEdge(a.id, b.id, 0.3 + Math.random() * 0.3, 'cross-domain')
      }
    }
  }

  return edges
}

function getCommunityList(nodes: GraphNode[]): Community[] {
  const map = new Map<number, { name: string; count: number }>()

  const names: Record<number, string> = {
    11: 'Foundations', 12: 'Attention Mechanism', 13: 'Training & Optimization',
    21: 'Large Language Models', 22: 'NLP Techniques',
    31: 'Vision Models',
    41: 'Generative Models', 42: 'Multimodal',
    51: 'RL Algorithms',
    61: 'Speech Models',
    71: 'AI Applications',
  }

  for (const node of nodes) {
    const existing = map.get(node.community_id)
    if (existing) {
      existing.count++
    } else {
      map.set(node.community_id, {
        name: names[node.community_id] || `Community ${node.community_id}`,
        count: 1,
      })
    }
  }

  return [...map.entries()].map(([id, info]) => ({
    id,
    name: info.name,
    color: LEVEL_PALETTES[1][id % LEVEL_PALETTES[1].length],
    node_count: info.count,
  }))
}

function buildHierarchicalCommunities(nodes: GraphNode[]): {
  root: HierarchicalCommunity
  allIds: Map<number, number[]>
} {
  const communityNames: Record<number, string> = {
    11: 'Foundations', 12: 'Attention Mechanism', 13: 'Training & Optimization',
    21: 'Large Language Models', 22: 'NLP Techniques',
    31: 'Vision Models',
    41: 'Generative Models', 42: 'Multimodal',
    51: 'RL Algorithms',
    61: 'Speech Models',
    71: 'AI Applications',
  }

  const totalNodes = nodes.length
  const l2Ids = [11, 12, 13, 21, 22, 31, 41, 42, 51, 61, 71]

  const allIds = new Map<number, number[]>()
  allIds.set(0, l2Ids)

  const countByL2 = new Map<number, number>()
  for (const node of nodes) {
    countByL2.set(node.community_id, (countByL2.get(node.community_id) || 0) + 1)
  }

  function getColor(level: number, idx: number) {
    const p = LEVEL_PALETTES[level] || LEVEL_PALETTES[1]
    return p[idx % p.length]
  }

  const l1Children: HierarchicalCommunity[] = [
    {
      id: 1, parent_id: 0, name: 'Deep Learning & Foundations', level: 1,
      color: getColor(1, 0),
      node_count: (countByL2.get(11) || 0) + (countByL2.get(12) || 0) + (countByL2.get(13) || 0),
      children: [
        { id: 11, parent_id: 1, name: communityNames[11], level: 2, color: getColor(2, 0), node_count: countByL2.get(11) || 0 },
        { id: 12, parent_id: 1, name: communityNames[12], level: 2, color: getColor(2, 1), node_count: countByL2.get(12) || 0 },
        { id: 13, parent_id: 1, name: communityNames[13], level: 2, color: getColor(2, 2), node_count: countByL2.get(13) || 0 },
      ],
    },
    {
      id: 2, parent_id: 0, name: 'Natural Language Processing', level: 1,
      color: getColor(1, 1),
      node_count: (countByL2.get(21) || 0) + (countByL2.get(22) || 0),
      children: [
        { id: 21, parent_id: 2, name: communityNames[21], level: 2, color: getColor(2, 3), node_count: countByL2.get(21) || 0 },
        { id: 22, parent_id: 2, name: communityNames[22], level: 2, color: getColor(2, 4), node_count: countByL2.get(22) || 0 },
      ],
    },
    {
      id: 3, parent_id: 0, name: 'Computer Vision', level: 1,
      color: getColor(1, 2),
      node_count: countByL2.get(31) || 0,
      children: [
        { id: 31, parent_id: 3, name: communityNames[31], level: 2, color: getColor(2, 5), node_count: countByL2.get(31) || 0 },
      ],
    },
    {
      id: 4, parent_id: 0, name: 'Generative Models & Multimodal', level: 1,
      color: getColor(1, 3),
      node_count: (countByL2.get(41) || 0) + (countByL2.get(42) || 0),
      children: [
        { id: 41, parent_id: 4, name: communityNames[41], level: 2, color: getColor(2, 6), node_count: countByL2.get(41) || 0 },
        { id: 42, parent_id: 4, name: communityNames[42], level: 2, color: getColor(2, 7), node_count: countByL2.get(42) || 0 },
      ],
    },
    {
      id: 5, parent_id: 0, name: 'Reinforcement Learning', level: 1,
      color: getColor(1, 4),
      node_count: countByL2.get(51) || 0,
      children: [
        { id: 51, parent_id: 5, name: communityNames[51], level: 2, color: getColor(2, 8), node_count: countByL2.get(51) || 0 },
      ],
    },
    {
      id: 6, parent_id: 0, name: 'Speech Processing', level: 1,
      color: getColor(1, 5),
      node_count: countByL2.get(61) || 0,
      children: [
        { id: 61, parent_id: 6, name: communityNames[61], level: 2, color: getColor(2, 9), node_count: countByL2.get(61) || 0 },
      ],
    },
    {
      id: 7, parent_id: 0, name: 'AI Applications', level: 1,
      color: getColor(1, 6),
      node_count: countByL2.get(71) || 0,
      children: [
        { id: 71, parent_id: 7, name: communityNames[71], level: 2, color: getColor(2, 10), node_count: countByL2.get(71) || 0 },
      ],
    },
  ]

  for (const child of l1Children) {
    allIds.set(child.id, child.children!.map(c => c.id))
  }

  for (const cid of l2Ids) {
    allIds.set(cid, [cid])
  }

  const root: HierarchicalCommunity = {
    id: 0, parent_id: null, name: 'All', level: 0,
    color: getColor(0, 0), node_count: totalNodes,
    children: l1Children,
  }

  return { root, allIds }
}
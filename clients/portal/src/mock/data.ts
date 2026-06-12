import type { GraphNode, GraphEdge, Community, HierarchicalCommunity, TimelineRange } from '../types/graph'
import { LEVEL_PALETTES } from '../types/graph'

export const rawNodes: (Omit<GraphNode, 'community_id' | 'degree'> & { first_appeared: string })[] = [
  // ══════════════════════════════════════════════
  // 深度学习基础架构
  // ══════════════════════════════════════════════
  { id: 'mlp', name: 'MLP', description: '多层感知机，神经网络的基本形态', influence_score: 8.0, first_appeared: '1957-01' },
  { id: 'cnn', name: 'CNN', description: '卷积神经网络，计算机视觉的基石', influence_score: 9.3, first_appeared: '1989-01' },
  { id: 'rnn', name: 'RNN', description: '循环神经网络，处理序列数据的基础架构', influence_score: 8.0, first_appeared: '1986-01' },
  { id: 'lstm', name: 'LSTM', description: '长短期记忆网络，解决长期依赖问题', influence_score: 8.5, first_appeared: '1997-01' },
  { id: 'autoencoder', name: 'Autoencoder', description: '自编码器，无监督表示学习经典架构', influence_score: 7.8, first_appeared: '1986-01' },
  { id: 'gru', name: 'GRU', description: '门控循环单元，LSTM的简化高效变体', influence_score: 7.5, first_appeared: '2014-01' },
  { id: 'encoder_decoder', name: 'Encoder-Decoder', description: '编码器-解码器架构，序列转换的基础范式', influence_score: 8.4, first_appeared: '2014-01' },
  { id: 'seq2seq', name: 'Seq2Seq', description: '序列到序列模型，机器翻译等任务的核心', influence_score: 8.0, first_appeared: '2014-01' },

  // ══════════════════════════════════════════════
  // Transformer 生态
  // ══════════════════════════════════════════════
  { id: 'transformer', name: 'Transformer', description: '基于自注意力的架构，彻底改变整个AI格局，论文引用超14万次', influence_score: 9.9, first_appeared: '2017-01' },
  { id: 'attention', name: 'Attention', description: '注意力机制，让模型学会选择性聚焦', influence_score: 9.5, first_appeared: '2014-01' },
  { id: 'self_attention', name: 'Self-Attention', description: '自注意力，序列内部元素之间的注意力计算', influence_score: 9.3, first_appeared: '2017-01' },
  { id: 'multi_head_attention', name: 'Multi-Head Attention', description: '多头注意力，并行捕获不同子空间特征', influence_score: 9.0, first_appeared: '2017-01' },
  { id: 'attention_mechanism', name: 'Attention Mechanism', description: '注意力机制的理论框架，选择性关注输入', influence_score: 9.2, first_appeared: '2014-01' },
  { id: 'positional_encoding', name: 'Positional Encoding', description: '位置编码，为Transformer注入序列位置信息', influence_score: 8.4, first_appeared: '2017-01' },
  { id: 'rope', name: 'RoPE', description: '旋转位置编码，LLaMA等主流模型采用的位置编码方案', influence_score: 8.2, first_appeared: '2021-01' },
  { id: 'gqa', name: 'Grouped Query Attention', description: '分组查询注意力，减少KV缓存内存占用', influence_score: 7.8, first_appeared: '2023-01' },
  { id: 'flash_attention', name: 'Flash Attention', description: '高效注意力计算，减少内存访问，加速大模型训练推理', influence_score: 8.6, first_appeared: '2022-01' },
  { id: 'kv_cache', name: 'KV Cache', description: '键值缓存，加速自回归推理', influence_score: 7.9, first_appeared: '2018-01' },
  { id: 'speculative_decoding', name: 'Speculative Decoding', description: '推测解码，用小模型加速大模型推理', influence_score: 7.5, first_appeared: '2023-01' },

  // ══════════════════════════════════════════════
  // 经典 NLP 模型
  // ══════════════════════════════════════════════
  { id: 'bert', name: 'BERT', description: '双向编码器表示模型，开启预训练微调范式', influence_score: 9.5, first_appeared: '2018-01' },
  { id: 'roberta', name: 'RoBERTa', description: 'BERT的优化版本，更鲁棒的预训练方法', influence_score: 8.5, first_appeared: '2019-01' },
  { id: 't5', name: 'T5', description: 'Text-to-Text Transfer Transformer统一框架', influence_score: 8.8, first_appeared: '2019-01' },

  // ══════════════════════════════════════════════
  // 大语言模型
  // ══════════════════════════════════════════════
  { id: 'llm', name: 'LLM', description: '大语言模型，大规模参数的预训练语言模型', influence_score: 9.6, first_appeared: '2020-01' },
  { id: 'gpt', name: 'GPT', description: '生成式预训练Transformer，开创大模型时代', influence_score: 9.4, first_appeared: '2018-01' },
  { id: 'gpt2', name: 'GPT-2', description: '15亿参数生成模型，因效果惊人一度不敢开源', influence_score: 8.7, first_appeared: '2019-01' },
  { id: 'gpt3', name: 'GPT-3', description: '1750亿参数大模型，展示涌现能力', influence_score: 9.3, first_appeared: '2020-01' },
  { id: 'gpt4', name: 'GPT-4', description: '多模态大语言模型，接近人类水平', influence_score: 9.7, first_appeared: '2023-01' },
  { id: 'chatgpt', name: 'ChatGPT', description: 'OpenAI对话式AI助手，两个月用户破亿', influence_score: 9.8, first_appeared: '2022-01' },
  { id: 'llama', name: 'LLaMA', description: 'Meta开源大语言模型，开启开源LLM浪潮', influence_score: 9.0, first_appeared: '2023-01' },
  { id: 'gemini', name: 'Gemini', description: 'Google的多模态大模型', influence_score: 9.0, first_appeared: '2023-01' },
  { id: 'claude', name: 'Claude', description: 'Anthropic的AI助手，以安全对齐著称', influence_score: 8.8, first_appeared: '2023-01' },
  { id: 'mistral', name: 'Mistral', description: 'Mistral AI的开源大模型', influence_score: 8.5, first_appeared: '2023-01' },
  { id: 'deepseek', name: 'DeepSeek', description: '深度求索大模型，高效开源模型', influence_score: 8.6, first_appeared: '2024-01' },
  { id: 'qwen', name: 'Qwen', description: '阿里通义千问大模型系列', influence_score: 8.3, first_appeared: '2023-01' },
  { id: 'grok', name: 'Grok', description: 'xAI的大语言模型', influence_score: 7.8, first_appeared: '2023-01' },
  { id: 'phi', name: 'Phi', description: '微软小参数高性能语言模型', influence_score: 8.0, first_appeared: '2023-01' },
  { id: 'yi', name: 'Yi', description: '零一万物大模型', influence_score: 7.8, first_appeared: '2023-01' },

  // ══════════════════════════════════════════════
  // NLP 技术
  // ══════════════════════════════════════════════
  { id: 'word2vec', name: 'Word2Vec', description: '词向量表示学习，推动NLP表示革命', influence_score: 8.6, first_appeared: '2013-01' },
  { id: 'glove', name: 'GloVe', description: '全局向量词表示', influence_score: 7.8, first_appeared: '2014-01' },
  { id: 'fasttext', name: 'fastText', description: 'Facebook的文本分类和子词词表示库', influence_score: 7.4, first_appeared: '2016-01' },
  { id: 'tokenization', name: 'Tokenization', description: '文本分词技术，将文本切分为模型输入单元', influence_score: 7.6, first_appeared: '1990-01' },
  { id: 'bpe', name: 'BPE', description: '字节对编码，子词分词算法', influence_score: 8.0, first_appeared: '2016-01' },
  { id: 'embedding', name: 'Embedding', description: '将离散符号映射到连续向量空间', influence_score: 8.7, first_appeared: '2013-01' },
  { id: 'prompt_engineering', name: 'Prompt Engineering', description: '提示工程，优化模型输入以改善输出', influence_score: 8.8, first_appeared: '2020-01' },
  { id: 'chain_of_thought', name: 'Chain-of-Thought', description: '思维链提示，引导模型逐步推理', influence_score: 9.0, first_appeared: '2022-01' },
  { id: 'rag', name: 'RAG', description: '检索增强生成，让LLM获取外部知识', influence_score: 9.0, first_appeared: '2023-01' },
  { id: 'in_context_learning', name: 'In-Context Learning', description: '上下文学习，无需参数更新的适应能力', influence_score: 8.9, first_appeared: '2020-01' },
  { id: 'function_calling', name: 'Function Calling', description: 'LLM调用外部工具/API的能力', influence_score: 8.4, first_appeared: '2023-01' },
  { id: 'tool_use', name: 'Tool Use', description: 'AI使用外部工具增强能力的范式', influence_score: 8.2, first_appeared: '2023-01' },
  { id: 'instruction_tuning', name: 'Instruction Tuning', description: '指令微调，让模型学会遵循人类指令', influence_score: 8.6, first_appeared: '2021-01' },
  { id: 'beam_search', name: 'Beam Search', description: '束搜索解码策略，生成更优序列', influence_score: 7.5, first_appeared: '2014-01' },
  { id: 'temperature', name: 'Temperature', description: '温度参数，控制生成输出的随机性和多样性', influence_score: 7.7, first_appeared: '2018-01' },
  { id: 'top_k', name: 'Top-K Sampling', description: 'Top-K采样策略，限制候选词范围', influence_score: 7.4, first_appeared: '2018-01' },
  { id: 'top_p', name: 'Top-P Sampling', description: '核采样策略，动态调整候选词集', influence_score: 7.5, first_appeared: '2019-01' },
  { id: 'hallucination', name: 'Hallucination', description: 'LLM生成不准确或虚构内容的现象与研究', influence_score: 8.4, first_appeared: '2020-01' },
  { id: 'softmax', name: 'Softmax', description: '将输出转化为概率分布的激活函数', influence_score: 8.2, first_appeared: '1989-01' },
  { id: 'cross_entropy', name: 'Cross Entropy', description: '交叉熵损失函数，分类任务的标准损失', influence_score: 8.0, first_appeared: '1948-01' },

  // ══════════════════════════════════════════════
  // 计算机视觉
  // ══════════════════════════════════════════════
  { id: 'resnet', name: 'ResNet', description: '残差网络，解决深度网络退化问题', influence_score: 9.3, first_appeared: '2015-01' },
  { id: 'vit', name: 'ViT', description: '视觉Transformer，将Transformer应用于图像', influence_score: 9.1, first_appeared: '2020-01' },
  { id: 'yolo', name: 'YOLO', description: '实时目标检测模型系列', influence_score: 8.7, first_appeared: '2015-01' },
  { id: 'faster_rcnn', name: 'Faster R-CNN', description: '基于区域提议网络的快速目标检测', influence_score: 8.3, first_appeared: '2015-01' },
  { id: 'mask_rcnn', name: 'Mask R-CNN', description: '实例分割模型，兼顾检测与分割', influence_score: 8.1, first_appeared: '2017-01' },
  { id: 'unet', name: 'U-Net', description: '编码器-解码器结构的语义分割网络', influence_score: 8.5, first_appeared: '2015-01' },
  { id: 'sam', name: 'SAM', description: 'Segment Anything Model，通用图像分割', influence_score: 8.8, first_appeared: '2023-01' },
  { id: 'detr', name: 'DETR', description: '基于Transformer的端到端目标检测模型', influence_score: 8.2, first_appeared: '2020-01' },
  { id: 'neRF', name: 'NeRF', description: '神经辐射场，3D场景重建', influence_score: 8.4, first_appeared: '2020-01' },
  { id: 'object_detection', name: 'Object Detection', description: '目标检测任务，识别图像中物体的类别和位置', influence_score: 8.2, first_appeared: '2012-01' },
  { id: 'image_segmentation', name: 'Image Segmentation', description: '图像分割，像素级分类', influence_score: 8.0, first_appeared: '2012-01' },
  { id: 'image_classification', name: 'Image Classification', description: '图像分类，判断图像所属类别', influence_score: 8.0, first_appeared: '2012-01' },
  { id: 'ocr', name: 'OCR', description: '光学字符识别', influence_score: 7.5, first_appeared: '1990-01' },
  { id: 'patch_embedding', name: 'Patch Embedding', description: '将图像分割为固定大小块并嵌入为向量', influence_score: 7.8, first_appeared: '2020-01' },
  { id: 'super_resolution', name: 'Super Resolution', description: '图像超分辨率重建', influence_score: 7.7, first_appeared: '2014-01' },
  { id: 'style_transfer', name: 'Style Transfer', description: '风格迁移，将艺术风格应用到图像', influence_score: 7.5, first_appeared: '2015-01' },

  // ══════════════════════════════════════════════
  // 生成模型
  // ══════════════════════════════════════════════
  { id: 'gan', name: 'GAN', description: '生成对抗网络，开创生成式AI新范式', influence_score: 9.0, first_appeared: '2014-01' },
  { id: 'vae', name: 'VAE', description: '变分自编码器，概率生成建模', influence_score: 8.1, first_appeared: '2013-01' },
  { id: 'diffusion_model', name: 'Diffusion Model', description: '扩散生成模型，逐步去噪生成高质量数据', influence_score: 9.2, first_appeared: '2015-01' },
  { id: 'ddpm', name: 'DDPM', description: '去噪扩散概率模型，扩散模型的基础算法', influence_score: 8.7, first_appeared: '2020-01' },
  { id: 'stable_diffusion', name: 'Stable Diffusion', description: '开源图像生成模型，推动AI创作民主化', influence_score: 9.4, first_appeared: '2022-01' },
  { id: 'stylegan', name: 'StyleGAN', description: '基于风格的生成对抗网络，高质量图像生成', influence_score: 8.6, first_appeared: '2018-01' },
  { id: 'midjourney', name: 'Midjourney', description: 'AI图像生成工具，美学质量领先', influence_score: 8.5, first_appeared: '2022-01' },
  { id: 'sora', name: 'Sora', description: 'OpenAI的文本到视频生成模型', influence_score: 9.0, first_appeared: '2024-01' },
  { id: 'dalle', name: 'DALL-E', description: '文本到图像生成模型', influence_score: 9.2, first_appeared: '2021-01' },
  { id: 'text_to_image', name: 'Text-to-Image', description: '文本到图像生成技术', influence_score: 8.8, first_appeared: '2021-01' },
  { id: 'image_inpainting', name: 'Image Inpainting', description: '图像修复，填充缺失区域', influence_score: 7.4, first_appeared: '2016-01' },
  { id: 'flow_matching', name: 'Flow Matching', description: '流匹配生成模型，扩散模型的替代方案', influence_score: 7.6, first_appeared: '2023-01' },
  { id: 'score_matching', name: 'Score Matching', description: '分数匹配，扩散模型的理论基础', influence_score: 7.5, first_appeared: '2015-01' },

  // ══════════════════════════════════════════════
  // 多模态
  // ══════════════════════════════════════════════
  { id: 'multimodal', name: 'Multimodal', description: '多模态AI，处理文本、图像、音频等多种输入', influence_score: 9.1, first_appeared: '2020-01' },
  { id: 'clip', name: 'CLIP', description: '对比语言-图像预训练模型', influence_score: 9.1, first_appeared: '2021-01' },
  { id: 'blip', name: 'BLIP', description: '引导语言图像预训练模型', influence_score: 8.4, first_appeared: '2022-01' },
  { id: 'flamingo', name: 'Flamingo', description: 'DeepMind的视觉语言模型', influence_score: 7.8, first_appeared: '2022-01' },
  { id: 'vision_language', name: 'Vision-Language', description: '视觉-语言模型，多模态子领域', influence_score: 8.6, first_appeared: '2021-01' },

  // ══════════════════════════════════════════════
  // 语音处理
  // ══════════════════════════════════════════════
  { id: 'whisper', name: 'Whisper', description: 'OpenAI的通用语音识别模型', influence_score: 8.3, first_appeared: '2022-01' },
  { id: 'wav2vec', name: 'wav2vec 2.0', description: '自监督语音表示学习框架', influence_score: 8.0, first_appeared: '2020-01' },
  { id: 'tts', name: 'TTS', description: '文本到语音合成', influence_score: 7.8, first_appeared: '2016-01' },
  { id: 'asr', name: 'ASR', description: '自动语音识别', influence_score: 7.9, first_appeared: '2006-01' },
  { id: 'music_generation', name: 'Music Generation', description: 'AI音乐生成', influence_score: 7.6, first_appeared: '2016-01' },

  // ══════════════════════════════════════════════
  // 强化学习
  // ══════════════════════════════════════════════
  { id: 'reinforcement_learning', name: 'Reinforcement Learning', description: '强化学习，通过奖励信号学习最优策略', influence_score: 8.8, first_appeared: '1989-01' },
  { id: 'dqn', name: 'DQN', description: '深度Q网络，将深度学习引入强化学习', influence_score: 8.0, first_appeared: '2013-01' },
  { id: 'ppo', name: 'PPO', description: '近端策略优化，稳定高效的策略梯度方法', influence_score: 8.3, first_appeared: '2017-01' },
  { id: 'alphago', name: 'AlphaGo', description: '围棋AI，击败人类世界冠军', influence_score: 9.0, first_appeared: '2016-01' },
  { id: 'q_learning', name: 'Q-Learning', description: '基于值函数的经典强化学习算法', influence_score: 7.8, first_appeared: '1989-01' },
  { id: 'policy_gradient', name: 'Policy Gradient', description: '策略梯度方法，直接优化策略', influence_score: 7.6, first_appeared: '1992-01' },
  { id: 'world_model', name: 'World Model', description: '世界模型，学习并模拟环境动态', influence_score: 7.6, first_appeared: '2018-01' },
  { id: 'imitation_learning', name: 'Imitation Learning', description: '模仿学习，从专家演示中学习', influence_score: 7.4, first_appeared: '2009-01' },
  { id: 'sac', name: 'SAC', description: '软演员-评论家算法，最大化策略熵', influence_score: 7.5, first_appeared: '2018-01' },
  { id: 'planning', name: 'Planning', description: 'AI规划能力，分解和解决复杂任务', influence_score: 7.8, first_appeared: '1971-01' },

  // ══════════════════════════════════════════════
  // 学习范式
  // ══════════════════════════════════════════════
  { id: 'supervised_learning', name: 'Supervised Learning', description: '监督学习，从标注数据中学习', influence_score: 8.5, first_appeared: '1959-01' },
  { id: 'unsupervised_learning', name: 'Unsupervised Learning', description: '无监督学习，从无标注数据中学习', influence_score: 8.0, first_appeared: '1980-01' },
  { id: 'self_supervised', name: 'Self-Supervised', description: '自监督学习，从数据自身构造监督信号', influence_score: 8.7, first_appeared: '2019-01' },
  { id: 'few_shot', name: 'Few-Shot', description: '少样本学习，仅需少量示例即可适应', influence_score: 8.1, first_appeared: '2019-01' },
  { id: 'zero_shot', name: 'Zero-Shot', description: '零样本学习，无需示例即可识别', influence_score: 7.9, first_appeared: '2018-01' },
  { id: 'contrastive_learning', name: 'Contrastive Learning', description: '对比学习，自监督表示学习重要范式', influence_score: 8.6, first_appeared: '2020-01' },
  { id: 'curriculum_learning', name: 'Curriculum Learning', description: '课程学习，从简单到困难逐步训练', influence_score: 7.3, first_appeared: '2009-01' },
  { id: 'meta_learning', name: 'Meta-Learning', description: '元学习，学会如何学习', influence_score: 7.6, first_appeared: '2017-01' },

  // ══════════════════════════════════════════════
  // 训练与优化
  // ══════════════════════════════════════════════
  { id: 'backpropagation', name: 'Backpropagation', description: '反向传播算法，训练所有深度神经网络的核心', influence_score: 9.8, first_appeared: '1986-01' },
  { id: 'adam', name: 'Adam', description: '自适应优化器，最广泛使用的深度学习优化器', influence_score: 9.0, first_appeared: '2014-01' },
  { id: 'sgd', name: 'SGD', description: '随机梯度下降，经典优化方法', influence_score: 8.2, first_appeared: '1951-01' },
  { id: 'gradient_descent', name: 'Gradient Descent', description: '梯度下降优化方法，一切优化算法的基础', influence_score: 8.8, first_appeared: '1951-01' },
  { id: 'learning_rate', name: 'Learning Rate', description: '学习率，控制参数更新步长的关键超参数', influence_score: 7.8, first_appeared: '1951-01' },
  { id: 'dropout', name: 'Dropout', description: '随机失活正则化，简单有效的防过拟合技术', influence_score: 8.5, first_appeared: '2012-01' },
];

export function getTimelineRange(): TimelineRange {
  const years = rawNodes.map(n => n.first_appeared ? parseInt(n.first_appeared.split('-')[0], 10) : null).filter((y): y is number => y != null)
  return {
    minYear: Math.min(...years),
    maxYear: Math.max(...years),
  }
}

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
    { source: 'transformer', target: 'bert', relation: '架构基础' },
    { source: 'transformer', target: 'gpt', relation: '架构基础' },
    { source: 'transformer', target: 'vit', relation: '应用于CV' },
    { source: 'transformer', target: 'attention', relation: '核心机制' },
    { source: 'transformer', target: 'self_attention', relation: '核心机制' },
    { source: 'transformer', target: 'multi_head_attention', relation: '核心机制' },
    { source: 'transformer', target: 'encoder_decoder', relation: '改进' },
    { source: 'transformer', target: 'seq2seq', relation: '取代' },
    { source: 'transformer', target: 'llm', relation: '架构基础' },
    { source: 'transformer', target: 'gpt3', relation: '架构基础' },
    { source: 'transformer', target: 'gpt4', relation: '架构基础' },
    { source: 'transformer', target: 'llama', relation: '架构基础' },
    { source: 'transformer', target: 'whisper', relation: '架构基础' },
    { source: 'transformer', target: 'detr', relation: '应用于CV' },

    { source: 'attention', target: 'bert', relation: '机制组成' },
    { source: 'attention', target: 'seq2seq', relation: '增强' },
    { source: 'attention', target: 'attention_mechanism', relation: '具体实现' },

    { source: 'bert', target: 'tokenization', relation: '依赖' },
    { source: 'bert', target: 'embedding', relation: '依赖' },
    { source: 'bert', target: 'encoder_decoder', relation: '仅用编码器' },
    { source: 'bert', target: 'roberta', relation: '优化改进' },
    { source: 'gpt', target: 'transformer', relation: '仅用解码器' },
    { source: 'gpt', target: 'embedding', relation: '依赖' },
    { source: 'gpt2', target: 'gpt', relation: '升级迭代' },
    { source: 'gpt3', target: 'gpt2', relation: '升级迭代' },
    { source: 'gpt3', target: 'in_context_learning', relation: '激发' },
    { source: 'gpt4', target: 'gpt3', relation: '升级迭代' },
    { source: 'gpt4', target: 'multimodal', relation: '实现' },
    { source: 'chatgpt', target: 'gpt3', relation: '基于' },
    { source: 'chatgpt', target: 'gpt4', relation: '基于' },
    { source: 'chatgpt', target: 'instruction_tuning', relation: '依赖' },
    { source: 'chatgpt', target: 'prompt_engineering', relation: '依赖' },
    { source: 'llm', target: 'chain_of_thought', relation: '推理增强' },
    { source: 'llm', target: 'rag', relation: '知识增强' },
    { source: 'llm', target: 'hallucination', relation: '挑战问题' },
    { source: 'llama', target: 'rope', relation: '使用' },
    { source: 'llama', target: 'gqa', relation: '使用' },
    { source: 'gemini', target: 'multimodal', relation: '实现' },
    { source: 'qwen', target: 'multimodal', relation: '扩展' },

    { source: 'word2vec', target: 'embedding', relation: '实现' },
    { source: 'glove', target: 'embedding', relation: '实现' },
    { source: 'chain_of_thought', target: 'gpt3', relation: '提升推理' },
    { source: 'rag', target: 'embedding', relation: '依赖检索' },
    { source: 'rag', target: 'llm', relation: '增强' },
    { source: 'instruction_tuning', target: 'supervised_learning', relation: '基于' },
    { source: 'function_calling', target: 'tool_use', relation: '实现方式' },
    { source: 'bpe', target: 'tokenization', relation: '实现' },
    { source: 'kv_cache', target: 'gpt', relation: '加速推理' },
    { source: 'flash_attention', target: 'self_attention', relation: '优化' },
    { source: 'flash_attention', target: 'gqa', relation: '可组合' },
    { source: 'speculative_decoding', target: 'llm', relation: '加速推理' },

    { source: 'cnn', target: 'resnet', relation: '基础架构' },
    { source: 'cnn', target: 'yolo', relation: '基础架构' },
    { source: 'cnn', target: 'faster_rcnn', relation: '基础架构' },
    { source: 'resnet', target: 'vit', relation: '被超越' },
    { source: 'vit', target: 'patch_embedding', relation: '依赖' },
    { source: 'vit', target: 'transformer', relation: '架构基础' },
    { source: 'vit', target: 'self_attention', relation: '使用' },
    { source: 'vit', target: 'image_classification', relation: '用于' },
    { source: 'detr', target: 'transformer', relation: '架构基础' },
    { source: 'detr', target: 'object_detection', relation: '用于' },
    { source: 'unet', target: 'cnn', relation: '基于' },
    { source: 'unet', target: 'stable_diffusion', relation: '骨架网络' },
    { source: 'sam', target: 'vit', relation: '图像编码' },
    { source: 'sam', target: 'image_segmentation', relation: '用于' },
    { source: 'mask_rcnn', target: 'faster_rcnn', relation: '扩展' },
    { source: 'mask_rcnn', target: 'image_segmentation', relation: '用于' },
    { source: 'yolo', target: 'object_detection', relation: '用于' },
    { source: 'neRF', target: 'mlp', relation: '使用' },
    { source: 'style_transfer', target: 'cnn', relation: '基于' },
    { source: 'super_resolution', target: 'cnn', relation: '基于' },
    { source: 'ocr', target: 'cnn', relation: '基于' },

    { source: 'gan', target: 'cnn', relation: '使用CNN' },
    { source: 'stylegan', target: 'gan', relation: '改进' },
    { source: 'diffusion_model', target: 'score_matching', relation: '理论基础' },
    { source: 'ddpm', target: 'diffusion_model', relation: '实现' },
    { source: 'stable_diffusion', target: 'ddpm', relation: '基于' },
    { source: 'stable_diffusion', target: 'clip', relation: '文本编码' },
    { source: 'stable_diffusion', target: 'text_to_image', relation: '实现' },
    { source: 'sora', target: 'diffusion_model', relation: '基于' },
    { source: 'sora', target: 'transformer', relation: '结合' },
    { source: 'dalle', target: 'clip', relation: '文本编码' },
    { source: 'dalle', target: 'gpt3', relation: '结合' },
    { source: 'dalle', target: 'text_to_image', relation: '实现' },
    { source: 'midjourney', target: 'diffusion_model', relation: '基于' },
    { source: 'midjourney', target: 'text_to_image', relation: '实现' },
    { source: 'flow_matching', target: 'diffusion_model', relation: '替代方案' },
    { source: 'vae', target: 'autoencoder', relation: '变体' },
    { source: 'vae', target: 'stable_diffusion', relation: '潜在空间' },
    { source: 'image_inpainting', target: 'gan', relation: '基于' },
    { source: 'image_inpainting', target: 'diffusion_model', relation: '基于' },

    { source: 'clip', target: 'contrastive_learning', relation: '基于' },
    { source: 'clip', target: 'vit', relation: '图像编码' },
    { source: 'clip', target: 'transformer', relation: '文本编码' },
    { source: 'clip', target: 'multimodal', relation: '基础模型' },
    { source: 'vision_language', target: 'multimodal', relation: '子领域' },
    { source: 'multimodal', target: 'vit', relation: '视觉模块' },
    { source: 'multimodal', target: 'transformer', relation: '融合模块' },
    { source: 'blip', target: 'clip', relation: '改进' },
    { source: 'flamingo', target: 'multimodal', relation: '实现' },

    { source: 'dqn', target: 'gradient_descent', relation: '优化' },
    { source: 'dqn', target: 'backpropagation', relation: '训练' },
    { source: 'ppo', target: 'policy_gradient', relation: '改进' },
    { source: 'alphago', target: 'reinforcement_learning', relation: '应用' },
    { source: 'alphago', target: 'dqn', relation: '结合' },
    { source: 'reinforcement_learning', target: 'q_learning', relation: '包含' },
    { source: 'reinforcement_learning', target: 'policy_gradient', relation: '包含' },
    { source: 'sac', target: 'policy_gradient', relation: '改进' },
    { source: 'imitation_learning', target: 'reinforcement_learning', relation: '范式' },
    { source: 'world_model', target: 'reinforcement_learning', relation: '扩展' },

    { source: 'whisper', target: 'transformer', relation: '架构基础' },
    { source: 'whisper', target: 'asr', relation: '实现' },
    { source: 'wav2vec', target: 'self_supervised', relation: '基于' },
    { source: 'wav2vec', target: 'transformer', relation: '使用' },
    { source: 'tts', target: 'transformer', relation: '基于' },
    { source: 'music_generation', target: 'transformer', relation: '基于' },

    { source: 'backpropagation', target: 'mlp', relation: '训练' },
    { source: 'backpropagation', target: 'cnn', relation: '训练' },
    { source: 'backpropagation', target: 'lstm', relation: '训练' },
    { source: 'adam', target: 'transformer', relation: '训练' },
    { source: 'adam', target: 'bert', relation: '训练' },
    { source: 'adam', target: 'gpt', relation: '训练' },
    { source: 'adam', target: 'resnet', relation: '训练' },
    { source: 'adam', target: 'stable_diffusion', relation: '训练' },
    { source: 'adam', target: 'clip', relation: '训练' },
    { source: 'dropout', target: 'transformer', relation: '正则化' },
    { source: 'dropout', target: 'bert', relation: '正则化' },
    { source: 'dropout', target: 'cnn', relation: '正则化' },
    { source: 'gradient_descent', target: 'sgd', relation: '基础' },
    { source: 'gradient_descent', target: 'backpropagation', relation: '结合' },
    { source: 'sgd', target: 'cnn', relation: '优化' },
    { source: 'learning_rate', target: 'adam', relation: '超参数' },
    { source: 'learning_rate', target: 'sgd', relation: '超参数' },

    { source: 'mlp', target: 'backpropagation', relation: '依赖' },
    { source: 'lstm', target: 'rnn', relation: '改进' },
    { source: 'lstm', target: 'encoder_decoder', relation: '组件' },
    { source: 'gru', target: 'lstm', relation: '简化变体' },
    { source: 'encoder_decoder', target: 'seq2seq', relation: '基础架构' },
    { source: 'seq2seq', target: 'attention', relation: '增强' },
    { source: 'rnn', target: 'lstm', relation: '变体基础' },
    { source: 'autoencoder', target: 'vae', relation: '变体' },

    { source: 'embedding', target: 'tokenization', relation: '后续步骤' },
    { source: 'embedding', target: 'softmax', relation: '输出层' },
    { source: 'cross_entropy', target: 'softmax', relation: '配对' },
    { source: 'self_supervised', target: 'bert', relation: '训练方式' },
    { source: 'self_supervised', target: 'contrastive_learning', relation: '包含' },
    { source: 'few_shot', target: 'gpt3', relation: '评估方式' },
    { source: 'few_shot', target: 'llm', relation: '能力' },
    { source: 'zero_shot', target: 'clip', relation: '能力' },
    { source: 'contrastive_learning', target: 'self_supervised', relation: '实现' },
    { source: 'contrastive_learning', target: 'simclr', relation: '方法' },
    { source: 'meta_learning', target: 'few_shot', relation: '方法' },
    { source: 'softmax', target: 'temperature', relation: '参数' },
    { source: 'top_k', target: 'temperature', relation: '采样' },
    { source: 'top_p', target: 'temperature', relation: '采样' },
    { source: 'beam_search', target: 'seq2seq', relation: '解码' },
    { source: 't5', target: 'encoder_decoder', relation: '架构' },
    { source: 't5', target: 'transformer', relation: '统一框架' },
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
        addEdge(n.id, shuffled[j].id, weight, '同领域')
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
        addEdge(a.id, b.id, 0.3 + Math.random() * 0.3, '跨领域')
      }
    }
  }

  return edges
}

function getCommunityList(nodes: GraphNode[]): Community[] {
  const map = new Map<number, { name: string; count: number }>()

  const names: Record<number, string> = {
    11: '基础架构', 12: '注意力机制', 13: '训练与优化',
    21: '大语言模型', 22: 'NLP 技术',
    31: '视觉模型',
    41: '生成模型', 42: '多模态',
    51: 'RL 算法',
    61: '语音模型',
    71: 'AI 应用',
  }

  for (const node of nodes) {
    const existing = map.get(node.community_id)
    if (existing) {
      existing.count++
    } else {
      map.set(node.community_id, {
        name: names[node.community_id] || `社区${node.community_id}`,
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
    11: '基础架构', 12: '注意力机制', 13: '训练与优化',
    21: '大语言模型', 22: 'NLP 技术',
    31: '视觉模型',
    41: '生成模型', 42: '多模态',
    51: 'RL 算法',
    61: '语音模型',
    71: 'AI 应用',
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
      id: 1, parent_id: 0, name: '深度学习与基础架构', level: 1,
      color: getColor(1, 0),
      node_count: (countByL2.get(11) || 0) + (countByL2.get(12) || 0) + (countByL2.get(13) || 0),
      children: [
        { id: 11, parent_id: 1, name: communityNames[11], level: 2, color: getColor(2, 0), node_count: countByL2.get(11) || 0 },
        { id: 12, parent_id: 1, name: communityNames[12], level: 2, color: getColor(2, 1), node_count: countByL2.get(12) || 0 },
        { id: 13, parent_id: 1, name: communityNames[13], level: 2, color: getColor(2, 2), node_count: countByL2.get(13) || 0 },
      ],
    },
    {
      id: 2, parent_id: 0, name: '自然语言处理', level: 1,
      color: getColor(1, 1),
      node_count: (countByL2.get(21) || 0) + (countByL2.get(22) || 0),
      children: [
        { id: 21, parent_id: 2, name: communityNames[21], level: 2, color: getColor(2, 3), node_count: countByL2.get(21) || 0 },
        { id: 22, parent_id: 2, name: communityNames[22], level: 2, color: getColor(2, 4), node_count: countByL2.get(22) || 0 },
      ],
    },
    {
      id: 3, parent_id: 0, name: '计算机视觉', level: 1,
      color: getColor(1, 2),
      node_count: countByL2.get(31) || 0,
      children: [
        { id: 31, parent_id: 3, name: communityNames[31], level: 2, color: getColor(2, 5), node_count: countByL2.get(31) || 0 },
      ],
    },
    {
      id: 4, parent_id: 0, name: '生成模型与多模态', level: 1,
      color: getColor(1, 3),
      node_count: (countByL2.get(41) || 0) + (countByL2.get(42) || 0),
      children: [
        { id: 41, parent_id: 4, name: communityNames[41], level: 2, color: getColor(2, 6), node_count: countByL2.get(41) || 0 },
        { id: 42, parent_id: 4, name: communityNames[42], level: 2, color: getColor(2, 7), node_count: countByL2.get(42) || 0 },
      ],
    },
    {
      id: 5, parent_id: 0, name: '强化学习', level: 1,
      color: getColor(1, 4),
      node_count: countByL2.get(51) || 0,
      children: [
        { id: 51, parent_id: 5, name: communityNames[51], level: 2, color: getColor(2, 8), node_count: countByL2.get(51) || 0 },
      ],
    },
    {
      id: 6, parent_id: 0, name: '语音处理', level: 1,
      color: getColor(1, 5),
      node_count: countByL2.get(61) || 0,
      children: [
        { id: 61, parent_id: 6, name: communityNames[61], level: 2, color: getColor(2, 9), node_count: countByL2.get(61) || 0 },
      ],
    },
    {
      id: 7, parent_id: 0, name: 'AI 应用', level: 1,
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
    id: 0, parent_id: null, name: '全部', level: 0,
    color: getColor(0, 0), node_count: totalNodes,
    children: l1Children,
  }

  return { root, allIds }
}
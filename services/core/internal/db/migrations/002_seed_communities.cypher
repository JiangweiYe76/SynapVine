-- Seed top-level communities for the AI domain
-- These are manually curated domain categories that LLM discovery will map into

MERGE (c:Community {id: 1, name: "深度学习架构", color: "#5a7a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 2, name: "NLP模型", color: "#6a8a6a", level: 1, domain: "ai"})
MERGE (c:Community {id: 3, name: "计算机视觉", color: "#7a6a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 4, name: "生成模型", color: "#8a6a5a", level: 1, domain: "ai"})
MERGE (c:Community {id: 5, name: "多模态", color: "#5a6a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 6, name: "语音模型", color: "#6a5a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 7, name: "图神经网络", color: "#8a7a5a", level: 1, domain: "ai"})
MERGE (c:Community {id: 8, name: "强化学习", color: "#5a8a7a", level: 1, domain: "ai"})
MERGE (c:Community {id: 9, name: "深度学习技术", color: "#7a8a5a", level: 1, domain: "ai"})
MERGE (c:Community {id: 10, name: "NLP技术", color: "#8a5a6a", level: 1, domain: "ai"})
MERGE (c:Community {id: 11, name: "优化器", color: "#6a7a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 12, name: "AI对齐", color: "#7a5a6a", level: 1, domain: "ai"})
MERGE (c:Community {id: 13, name: "平台与基础设施", color: "#5a6a6a", level: 1, domain: "ai"})
MERGE (c:Community {id: 14, name: "应用", color: "#4a5a6a", level: 1, domain: "ai"});

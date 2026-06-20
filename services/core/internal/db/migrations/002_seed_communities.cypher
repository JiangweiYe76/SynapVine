-- Seed top-level communities for the AI domain
-- These are manually curated domain categories that LLM discovery will map into

MERGE (c:Community {id: 1, name: "Deep Learning Architecture", color: "#5a7a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 2, name: "NLP Models", color: "#6a8a6a", level: 1, domain: "ai"})
MERGE (c:Community {id: 3, name: "Computer Vision", color: "#7a6a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 4, name: "Generative Models", color: "#8a6a5a", level: 1, domain: "ai"})
MERGE (c:Community {id: 5, name: "Multimodal", color: "#5a6a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 6, name: "Speech Models", color: "#6a5a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 7, name: "Graph Neural Networks", color: "#8a7a5a", level: 1, domain: "ai"})
MERGE (c:Community {id: 8, name: "Reinforcement Learning", color: "#5a8a7a", level: 1, domain: "ai"})
MERGE (c:Community {id: 9, name: "Deep Learning Techniques", color: "#7a8a5a", level: 1, domain: "ai"})
MERGE (c:Community {id: 10, name: "NLP Techniques", color: "#8a5a6a", level: 1, domain: "ai"})
MERGE (c:Community {id: 11, name: "Optimizers", color: "#6a7a8a", level: 1, domain: "ai"})
MERGE (c:Community {id: 12, name: "AI Alignment", color: "#7a5a6a", level: 1, domain: "ai"})
MERGE (c:Community {id: 13, name: "Platforms & Infrastructure", color: "#5a6a6a", level: 1, domain: "ai"})
MERGE (c:Community {id: 14, name: "Applications", color: "#4a5a6a", level: 1, domain: "ai"});

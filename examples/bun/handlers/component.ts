import { BUN_INTERACTION_PREFIX, ComponentHandler } from "./common.ts";
import {
    APIInteraction,
    APIMessageComponentInteraction,
    InteractionType,
} from "discord-api-types/v10";
import { counterComponentHandler } from "./commands/counter.ts";

export const componentHandlers: Record<string, ComponentHandler> = {
    counter: counterComponentHandler,
};

export const isComponent = (
    interaction: APIInteraction,
): interaction is APIMessageComponentInteraction => {
    if (interaction?.type !== InteractionType.MessageComponent) {
        return false;
    }
    const parts = interaction?.data?.custom_id?.split("/");
    if (!parts) {
        return false;
    }
    if (parts[0] !== BUN_INTERACTION_PREFIX) {
        return false;
    }
    return parts[1] in componentHandlers;
};

export const executeComponent: ComponentHandler = (interaction) =>
    componentHandlers[interaction?.data?.custom_id?.split("/")[1]](interaction);

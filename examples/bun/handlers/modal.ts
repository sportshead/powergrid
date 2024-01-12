import { BUN_INTERACTION_PREFIX, ModalHandler } from "./common.ts";
import {
    APIInteraction,
    APIModalSubmitInteraction,
    InteractionType,
} from "discord-api-types/v10";
import { counterModalHandler } from "./commands/counter.ts";

export const modalHandlers: Record<string, ModalHandler> = {
    counter: counterModalHandler,
};

export const isModal = (
    interaction: APIInteraction,
): interaction is APIModalSubmitInteraction => {
    if (interaction?.type !== InteractionType.ModalSubmit) {
        return false;
    }
    const parts = interaction?.data?.custom_id?.split("/");
    if (!parts) {
        return false;
    }
    if (parts[0] !== BUN_INTERACTION_PREFIX) {
        return false;
    }
    return parts[1] in modalHandlers;
};

export const executeModal: ModalHandler = (interaction) =>
    modalHandlers[interaction?.data?.custom_id?.split("/")[1]](interaction);

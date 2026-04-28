import { ref } from 'vue';
import { useAppI18n } from '@/i18n';
import { createPluginMenuNode } from '@/utils/plugin';
defineOptions({ name: 'PluginMenuTreeEditor' });
const emit = defineEmits();
const props = defineProps();
const { t } = useAppI18n();
const mutableMenus = props.menus;
const draggingId = ref('');
const dropHint = ref(null);
function addRootMenu() {
    mutableMenus.push(createPluginMenuNode(props.pluginName));
}
function addChildMenu(menu) {
    menu.children = menu.children ?? [];
    menu.children.push(createPluginMenuNode(props.pluginName, menu.id));
}
function removeMenu(list, index) {
    list.splice(index, 1);
}
function forwardMove(sourceId, targetId, position) {
    emit('move-node', sourceId, targetId, position);
}
function onDragStart(event, menu) {
    draggingId.value = menu.id;
    if (event.dataTransfer) {
        event.dataTransfer.effectAllowed = 'move';
        event.dataTransfer.setData('text/plain', menu.id);
    }
}
function onDragEnd() {
    draggingId.value = '';
    dropHint.value = null;
}
function setDropHint(targetId, position) {
    dropHint.value = { targetId, position };
}
function clearDropHint() {
    dropHint.value = null;
}
function isDropHint(targetId, position) {
    return dropHint.value?.targetId === targetId && dropHint.value?.position === position;
}
function getMenuDisplayTitle(menu) {
    return t(menu.titleKey || '', menu.titleDefault || menu.name || menu.id || t('plugin.menu_unnamed', 'Unnamed menu'));
}
function onDrop(event, targetId, position) {
    event.preventDefault();
    const sourceId = event.dataTransfer?.getData('text/plain') || draggingId.value;
    clearDropHint();
    if (!sourceId || sourceId === targetId) {
        return;
    }
    emit('move-node', sourceId, targetId, position);
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "plugin-menu-tree-editor" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "admin-table__actions mb-12" },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (__VLS_ctx.addRootMenu)
};
__VLS_3.slots.default;
(__VLS_ctx.t('plugin.add_root_menu', 'Add root menu'));
var __VLS_3;
if (__VLS_ctx.mutableMenus.length === 0) {
    const __VLS_8 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
        description: (__VLS_ctx.t('plugin.no_menus', 'No menus yet, please add a root menu first')),
    }));
    const __VLS_10 = __VLS_9({
        description: (__VLS_ctx.t('plugin.no_menus', 'No menus yet, please add a root menu first')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_9));
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "plugin-menu-tree-editor__list" },
    });
    for (const [menu, index] of __VLS_getVForSourceType((__VLS_ctx.mutableMenus))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (menu.id),
            ...{ class: "plugin-menu-tree-editor__node" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onDragover: (...[$event]) => {
                    if (!!(__VLS_ctx.mutableMenus.length === 0))
                        return;
                    __VLS_ctx.setDropHint(menu.id, 'before');
                } },
            ...{ onDragleave: (__VLS_ctx.clearDropHint) },
            ...{ onDrop: (...[$event]) => {
                    if (!!(__VLS_ctx.mutableMenus.length === 0))
                        return;
                    __VLS_ctx.onDrop($event, menu.id, 'before');
                } },
            ...{ class: "plugin-menu-tree-editor__dropzone" },
            ...{ class: ({ 'is-active': __VLS_ctx.isDropHint(menu.id, 'before') }) },
        });
        (__VLS_ctx.t('plugin.drop_before', 'Drop here to place before {name}', { name: menu.name || menu.id || __VLS_ctx.t('plugin.current_menu', 'Current menu') }));
        const __VLS_12 = {}.ElCard;
        /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
        // @ts-ignore
        const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
            ...{ 'onDragstart': {} },
            ...{ 'onDragend': {} },
            shadow: "never",
            ...{ class: "plugin-menu-tree-editor__card" },
            draggable: "true",
        }));
        const __VLS_14 = __VLS_13({
            ...{ 'onDragstart': {} },
            ...{ 'onDragend': {} },
            shadow: "never",
            ...{ class: "plugin-menu-tree-editor__card" },
            draggable: "true",
        }, ...__VLS_functionalComponentArgsRest(__VLS_13));
        let __VLS_16;
        let __VLS_17;
        let __VLS_18;
        const __VLS_19 = {
            onDragstart: (...[$event]) => {
                if (!!(__VLS_ctx.mutableMenus.length === 0))
                    return;
                __VLS_ctx.onDragStart($event, menu);
            }
        };
        const __VLS_20 = {
            onDragend: (__VLS_ctx.onDragEnd)
        };
        __VLS_15.slots.default;
        {
            const { header: __VLS_thisSlot } = __VLS_15.slots;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "page-card__header" },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            (__VLS_ctx.getMenuDisplayTitle(menu));
            const __VLS_21 = {}.ElSpace;
            /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
            // @ts-ignore
            const __VLS_22 = __VLS_asFunctionalComponent(__VLS_21, new __VLS_21({
                wrap: true,
            }));
            const __VLS_23 = __VLS_22({
                wrap: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_22));
            __VLS_24.slots.default;
            const __VLS_25 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_26 = __VLS_asFunctionalComponent(__VLS_25, new __VLS_25({
                effect: "plain",
            }));
            const __VLS_27 = __VLS_26({
                effect: "plain",
            }, ...__VLS_functionalComponentArgsRest(__VLS_26));
            __VLS_28.slots.default;
            (menu.type || __VLS_ctx.t('plugin.menu_type_menu', 'Menu'));
            var __VLS_28;
            const __VLS_29 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_30 = __VLS_asFunctionalComponent(__VLS_29, new __VLS_29({
                effect: "plain",
                type: "info",
            }));
            const __VLS_31 = __VLS_30({
                effect: "plain",
                type: "info",
            }, ...__VLS_functionalComponentArgsRest(__VLS_30));
            __VLS_32.slots.default;
            (__VLS_ctx.t('plugin.drag_sorting', 'Drag sorting'));
            var __VLS_32;
            const __VLS_33 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({
                ...{ 'onClick': {} },
                size: "small",
                type: "primary",
                plain: true,
            }));
            const __VLS_35 = __VLS_34({
                ...{ 'onClick': {} },
                size: "small",
                type: "primary",
                plain: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_34));
            let __VLS_37;
            let __VLS_38;
            let __VLS_39;
            const __VLS_40 = {
                onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.mutableMenus.length === 0))
                        return;
                    __VLS_ctx.addChildMenu(menu);
                }
            };
            __VLS_36.slots.default;
            (__VLS_ctx.t('plugin.add_child_menu', 'Add child menu'));
            var __VLS_36;
            const __VLS_41 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({
                ...{ 'onClick': {} },
                size: "small",
                type: "danger",
                plain: true,
            }));
            const __VLS_43 = __VLS_42({
                ...{ 'onClick': {} },
                size: "small",
                type: "danger",
                plain: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_42));
            let __VLS_45;
            let __VLS_46;
            let __VLS_47;
            const __VLS_48 = {
                onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.mutableMenus.length === 0))
                        return;
                    __VLS_ctx.removeMenu(__VLS_ctx.mutableMenus, index);
                }
            };
            __VLS_44.slots.default;
            (__VLS_ctx.t('common.delete', 'Delete'));
            var __VLS_44;
            var __VLS_24;
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onDragover: (...[$event]) => {
                    if (!!(__VLS_ctx.mutableMenus.length === 0))
                        return;
                    __VLS_ctx.setDropHint(menu.id, 'inside');
                } },
            ...{ onDragleave: (__VLS_ctx.clearDropHint) },
            ...{ onDrop: (...[$event]) => {
                    if (!!(__VLS_ctx.mutableMenus.length === 0))
                        return;
                    __VLS_ctx.onDrop($event, menu.id, 'inside');
                } },
            ...{ class: "plugin-menu-tree-editor__dropzone plugin-menu-tree-editor__dropzone--inside" },
            ...{ class: ({ 'is-active': __VLS_ctx.isDropHint(menu.id, 'inside') }) },
        });
        (__VLS_ctx.t('plugin.drop_inside', 'Drop here to place as a child of {name}', { name: menu.name || menu.id || __VLS_ctx.t('plugin.current_menu', 'Current menu') }));
        const __VLS_49 = {}.ElForm;
        /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
        // @ts-ignore
        const __VLS_50 = __VLS_asFunctionalComponent(__VLS_49, new __VLS_49({
            labelWidth: "96px",
            ...{ class: "admin-form admin-form--two-col" },
        }));
        const __VLS_51 = __VLS_50({
            labelWidth: "96px",
            ...{ class: "admin-form admin-form--two-col" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_50));
        __VLS_52.slots.default;
        const __VLS_53 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_54 = __VLS_asFunctionalComponent(__VLS_53, new __VLS_53({
            label: (__VLS_ctx.t('plugin.menu_id', 'Menu ID')),
        }));
        const __VLS_55 = __VLS_54({
            label: (__VLS_ctx.t('plugin.menu_id', 'Menu ID')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_54));
        __VLS_56.slots.default;
        const __VLS_57 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({
            modelValue: (menu.id),
            placeholder: (__VLS_ctx.t('plugin.menu_id_placeholder', 'Unique ID')),
        }));
        const __VLS_59 = __VLS_58({
            modelValue: (menu.id),
            placeholder: (__VLS_ctx.t('plugin.menu_id_placeholder', 'Unique ID')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_58));
        var __VLS_56;
        const __VLS_61 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({
            label: (__VLS_ctx.t('plugin.parent_id', 'Parent ID')),
        }));
        const __VLS_63 = __VLS_62({
            label: (__VLS_ctx.t('plugin.parent_id', 'Parent ID')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_62));
        __VLS_64.slots.default;
        const __VLS_65 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_66 = __VLS_asFunctionalComponent(__VLS_65, new __VLS_65({
            modelValue: (menu.parent_id),
            placeholder: (__VLS_ctx.t('plugin.parent_id_placeholder', 'Parent menu ID')),
        }));
        const __VLS_67 = __VLS_66({
            modelValue: (menu.parent_id),
            placeholder: (__VLS_ctx.t('plugin.parent_id_placeholder', 'Parent menu ID')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_66));
        var __VLS_64;
        const __VLS_69 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({
            label: (__VLS_ctx.t('plugin.menu_name', 'Name')),
            required: true,
        }));
        const __VLS_71 = __VLS_70({
            label: (__VLS_ctx.t('plugin.menu_name', 'Name')),
            required: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_70));
        __VLS_72.slots.default;
        const __VLS_73 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
            modelValue: (menu.name),
            placeholder: (__VLS_ctx.t('plugin.menu_name_placeholder', 'Menu name')),
        }));
        const __VLS_75 = __VLS_74({
            modelValue: (menu.name),
            placeholder: (__VLS_ctx.t('plugin.menu_name_placeholder', 'Menu name')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_74));
        var __VLS_72;
        const __VLS_77 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
            label: (__VLS_ctx.t('plugin.menu_title_key', 'Title key')),
        }));
        const __VLS_79 = __VLS_78({
            label: (__VLS_ctx.t('plugin.menu_title_key', 'Title key')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_78));
        __VLS_80.slots.default;
        const __VLS_81 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
            modelValue: (menu.titleKey),
            placeholder: (__VLS_ctx.t('plugin.menu_title_key_placeholder', 'For example, route.dashboard')),
        }));
        const __VLS_83 = __VLS_82({
            modelValue: (menu.titleKey),
            placeholder: (__VLS_ctx.t('plugin.menu_title_key_placeholder', 'For example, route.dashboard')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_82));
        var __VLS_80;
        const __VLS_85 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
            label: (__VLS_ctx.t('plugin.menu_title_default', 'Default title')),
        }));
        const __VLS_87 = __VLS_86({
            label: (__VLS_ctx.t('plugin.menu_title_default', 'Default title')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_86));
        __VLS_88.slots.default;
        const __VLS_89 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
            modelValue: (menu.titleDefault),
            placeholder: (__VLS_ctx.t('plugin.menu_title_default_placeholder', 'For example, Dashboard')),
        }));
        const __VLS_91 = __VLS_90({
            modelValue: (menu.titleDefault),
            placeholder: (__VLS_ctx.t('plugin.menu_title_default_placeholder', 'For example, Dashboard')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_90));
        var __VLS_88;
        const __VLS_93 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
            label: (__VLS_ctx.t('plugin.path', 'Path')),
            required: true,
        }));
        const __VLS_95 = __VLS_94({
            label: (__VLS_ctx.t('plugin.path', 'Path')),
            required: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_94));
        __VLS_96.slots.default;
        const __VLS_97 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
            modelValue: (menu.path),
            placeholder: (__VLS_ctx.t('plugin.menu_path_placeholder', '/plugin/example/home')),
        }));
        const __VLS_99 = __VLS_98({
            modelValue: (menu.path),
            placeholder: (__VLS_ctx.t('plugin.menu_path_placeholder', '/plugin/example/home')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_98));
        var __VLS_96;
        const __VLS_101 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
            label: (__VLS_ctx.t('plugin.component', 'Component')),
        }));
        const __VLS_103 = __VLS_102({
            label: (__VLS_ctx.t('plugin.component', 'Component')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_102));
        __VLS_104.slots.default;
        const __VLS_105 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({
            modelValue: (menu.component),
            placeholder: (__VLS_ctx.t('plugin.menu_component_placeholder_detail', 'view/plugin/example/index')),
        }));
        const __VLS_107 = __VLS_106({
            modelValue: (menu.component),
            placeholder: (__VLS_ctx.t('plugin.menu_component_placeholder_detail', 'view/plugin/example/index')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_106));
        var __VLS_104;
        const __VLS_109 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_110 = __VLS_asFunctionalComponent(__VLS_109, new __VLS_109({
            label: (__VLS_ctx.t('plugin.icon', 'Icon')),
        }));
        const __VLS_111 = __VLS_110({
            label: (__VLS_ctx.t('plugin.icon', 'Icon')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_110));
        __VLS_112.slots.default;
        const __VLS_113 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_114 = __VLS_asFunctionalComponent(__VLS_113, new __VLS_113({
            modelValue: (menu.icon),
            placeholder: (__VLS_ctx.t('plugin.icon_placeholder', 'sparkles')),
        }));
        const __VLS_115 = __VLS_114({
            modelValue: (menu.icon),
            placeholder: (__VLS_ctx.t('plugin.icon_placeholder', 'sparkles')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_114));
        var __VLS_112;
        const __VLS_117 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_118 = __VLS_asFunctionalComponent(__VLS_117, new __VLS_117({
            label: (__VLS_ctx.t('plugin.permission_key', 'Permission key')),
        }));
        const __VLS_119 = __VLS_118({
            label: (__VLS_ctx.t('plugin.permission_key', 'Permission key')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_118));
        __VLS_120.slots.default;
        const __VLS_121 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_122 = __VLS_asFunctionalComponent(__VLS_121, new __VLS_121({
            modelValue: (menu.permission),
            placeholder: (__VLS_ctx.t('plugin.permission_key_placeholder', 'plugin:example:view')),
        }));
        const __VLS_123 = __VLS_122({
            modelValue: (menu.permission),
            placeholder: (__VLS_ctx.t('plugin.permission_key_placeholder', 'plugin:example:view')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_122));
        var __VLS_120;
        const __VLS_125 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_126 = __VLS_asFunctionalComponent(__VLS_125, new __VLS_125({
            label: (__VLS_ctx.t('plugin.type', 'Type')),
        }));
        const __VLS_127 = __VLS_126({
            label: (__VLS_ctx.t('plugin.type', 'Type')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_126));
        __VLS_128.slots.default;
        const __VLS_129 = {}.ElSelect;
        /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
        // @ts-ignore
        const __VLS_130 = __VLS_asFunctionalComponent(__VLS_129, new __VLS_129({
            modelValue: (menu.type),
            ...{ style: {} },
        }));
        const __VLS_131 = __VLS_130({
            modelValue: (menu.type),
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_130));
        __VLS_132.slots.default;
        const __VLS_133 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_134 = __VLS_asFunctionalComponent(__VLS_133, new __VLS_133({
            label: (__VLS_ctx.t('plugin.menu_type_directory', 'Directory')),
            value: "directory",
        }));
        const __VLS_135 = __VLS_134({
            label: (__VLS_ctx.t('plugin.menu_type_directory', 'Directory')),
            value: "directory",
        }, ...__VLS_functionalComponentArgsRest(__VLS_134));
        const __VLS_137 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_138 = __VLS_asFunctionalComponent(__VLS_137, new __VLS_137({
            label: (__VLS_ctx.t('plugin.menu_type_menu', 'Menu')),
            value: "menu",
        }));
        const __VLS_139 = __VLS_138({
            label: (__VLS_ctx.t('plugin.menu_type_menu', 'Menu')),
            value: "menu",
        }, ...__VLS_functionalComponentArgsRest(__VLS_138));
        const __VLS_141 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_142 = __VLS_asFunctionalComponent(__VLS_141, new __VLS_141({
            label: (__VLS_ctx.t('plugin.menu_type_button', 'Button')),
            value: "button",
        }));
        const __VLS_143 = __VLS_142({
            label: (__VLS_ctx.t('plugin.menu_type_button', 'Button')),
            value: "button",
        }, ...__VLS_functionalComponentArgsRest(__VLS_142));
        var __VLS_132;
        var __VLS_128;
        const __VLS_145 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_146 = __VLS_asFunctionalComponent(__VLS_145, new __VLS_145({
            label: (__VLS_ctx.t('plugin.sort', 'Sort')),
        }));
        const __VLS_147 = __VLS_146({
            label: (__VLS_ctx.t('plugin.sort', 'Sort')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_146));
        __VLS_148.slots.default;
        const __VLS_149 = {}.ElInputNumber;
        /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
        // @ts-ignore
        const __VLS_150 = __VLS_asFunctionalComponent(__VLS_149, new __VLS_149({
            modelValue: (menu.sort),
            min: (0),
            step: (1),
            ...{ style: {} },
        }));
        const __VLS_151 = __VLS_150({
            modelValue: (menu.sort),
            min: (0),
            step: (1),
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_150));
        var __VLS_148;
        const __VLS_153 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_154 = __VLS_asFunctionalComponent(__VLS_153, new __VLS_153({
            label: (__VLS_ctx.t('plugin.redirect', 'Redirect')),
        }));
        const __VLS_155 = __VLS_154({
            label: (__VLS_ctx.t('plugin.redirect', 'Redirect')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_154));
        __VLS_156.slots.default;
        const __VLS_157 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_158 = __VLS_asFunctionalComponent(__VLS_157, new __VLS_157({
            modelValue: (menu.redirect),
            placeholder: (__VLS_ctx.t('plugin.redirect_placeholder', '/plugin/example/home')),
        }));
        const __VLS_159 = __VLS_158({
            modelValue: (menu.redirect),
            placeholder: (__VLS_ctx.t('plugin.redirect_placeholder', '/plugin/example/home')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_158));
        var __VLS_156;
        const __VLS_161 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_162 = __VLS_asFunctionalComponent(__VLS_161, new __VLS_161({
            label: (__VLS_ctx.t('plugin.external_url', 'External URL')),
        }));
        const __VLS_163 = __VLS_162({
            label: (__VLS_ctx.t('plugin.external_url', 'External URL')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_162));
        __VLS_164.slots.default;
        const __VLS_165 = {}.ElInput;
        /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
        // @ts-ignore
        const __VLS_166 = __VLS_asFunctionalComponent(__VLS_165, new __VLS_165({
            modelValue: (menu.external_url),
            placeholder: (__VLS_ctx.t('plugin.optional', 'Optional')),
        }));
        const __VLS_167 = __VLS_166({
            modelValue: (menu.external_url),
            placeholder: (__VLS_ctx.t('plugin.optional', 'Optional')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_166));
        var __VLS_164;
        const __VLS_169 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_170 = __VLS_asFunctionalComponent(__VLS_169, new __VLS_169({
            label: (__VLS_ctx.t('plugin.visible', 'Visible')),
        }));
        const __VLS_171 = __VLS_170({
            label: (__VLS_ctx.t('plugin.visible', 'Visible')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_170));
        __VLS_172.slots.default;
        const __VLS_173 = {}.ElSwitch;
        /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
        // @ts-ignore
        const __VLS_174 = __VLS_asFunctionalComponent(__VLS_173, new __VLS_173({
            modelValue: (menu.visible),
        }));
        const __VLS_175 = __VLS_174({
            modelValue: (menu.visible),
        }, ...__VLS_functionalComponentArgsRest(__VLS_174));
        var __VLS_172;
        const __VLS_177 = {}.ElFormItem;
        /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
        // @ts-ignore
        const __VLS_178 = __VLS_asFunctionalComponent(__VLS_177, new __VLS_177({
            label: (__VLS_ctx.t('plugin.enabled', 'Enabled')),
        }));
        const __VLS_179 = __VLS_178({
            label: (__VLS_ctx.t('plugin.enabled', 'Enabled')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_178));
        __VLS_180.slots.default;
        const __VLS_181 = {}.ElSwitch;
        /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
        // @ts-ignore
        const __VLS_182 = __VLS_asFunctionalComponent(__VLS_181, new __VLS_181({
            modelValue: (menu.enabled),
        }));
        const __VLS_183 = __VLS_182({
            modelValue: (menu.enabled),
        }, ...__VLS_functionalComponentArgsRest(__VLS_182));
        var __VLS_180;
        var __VLS_52;
        if (menu.children && menu.children.length > 0) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "plugin-menu-tree-editor__children" },
            });
            const __VLS_185 = {}.PluginMenuTreeEditor;
            /** @type {[typeof __VLS_components.PluginMenuTreeEditor, ]} */ ;
            // @ts-ignore
            const __VLS_186 = __VLS_asFunctionalComponent(__VLS_185, new __VLS_185({
                ...{ 'onMoveNode': {} },
                menus: (menu.children),
                pluginName: (__VLS_ctx.pluginName),
            }));
            const __VLS_187 = __VLS_186({
                ...{ 'onMoveNode': {} },
                menus: (menu.children),
                pluginName: (__VLS_ctx.pluginName),
            }, ...__VLS_functionalComponentArgsRest(__VLS_186));
            let __VLS_189;
            let __VLS_190;
            let __VLS_191;
            const __VLS_192 = {
                onMoveNode: (__VLS_ctx.forwardMove)
            };
            var __VLS_188;
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onDragover: (...[$event]) => {
                    if (!!(__VLS_ctx.mutableMenus.length === 0))
                        return;
                    __VLS_ctx.setDropHint(menu.id, 'after');
                } },
            ...{ onDragleave: (__VLS_ctx.clearDropHint) },
            ...{ onDrop: (...[$event]) => {
                    if (!!(__VLS_ctx.mutableMenus.length === 0))
                        return;
                    __VLS_ctx.onDrop($event, menu.id, 'after');
                } },
            ...{ class: "plugin-menu-tree-editor__dropzone" },
            ...{ class: ({ 'is-active': __VLS_ctx.isDropHint(menu.id, 'after') }) },
        });
        (__VLS_ctx.t('plugin.drop_after', 'Drop here to place after {name}', { name: __VLS_ctx.getMenuDisplayTitle(menu) }));
        var __VLS_15;
    }
}
/** @type {__VLS_StyleScopedClasses['plugin-menu-tree-editor']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-table__actions']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-menu-tree-editor__list']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-menu-tree-editor__node']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-menu-tree-editor__dropzone']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-menu-tree-editor__card']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-menu-tree-editor__dropzone']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-menu-tree-editor__dropzone--inside']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form--two-col']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-menu-tree-editor__children']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-menu-tree-editor__dropzone']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            mutableMenus: mutableMenus,
            addRootMenu: addRootMenu,
            addChildMenu: addChildMenu,
            removeMenu: removeMenu,
            forwardMove: forwardMove,
            onDragStart: onDragStart,
            onDragEnd: onDragEnd,
            setDropHint: setDropHint,
            clearDropHint: clearDropHint,
            isDropHint: isDropHint,
            getMenuDisplayTitle: getMenuDisplayTitle,
            onDrop: onDrop,
        };
    },
    __typeEmits: {},
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */

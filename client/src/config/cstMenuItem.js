import { ARTICLE, HOME, DEMO, PROJECT } from "./cstModule"
import { MenuOutlined } from "@ant-design/icons"
import React from "react"

export const MENU_ITEMS_WITHOUT_SUB = [
    {
        key: HOME.KEY,
        className: "menu-item",
        label: HOME.NAME,
        icon: null,
    },
    {
        key: ARTICLE.KEY,
        className: "menu-item",
        label: ARTICLE.NAME,
        icon: null,
    },
    {
        key: DEMO.KEY,
        className: "menu-item",
        label: DEMO.NAME,
        icon: null,
    },
    {
        key: PROJECT.KEY,
        className: "menu-item",
        label: PROJECT.NAME,
        icon: null,
    },
]

export const MENU_ITEMS_WITH_SUB = [
    {
        label: "",
        key: "submenu",
        icon: <MenuOutlined className="submenu-icon" />,
        children: MENU_ITEMS_WITHOUT_SUB
    },
]
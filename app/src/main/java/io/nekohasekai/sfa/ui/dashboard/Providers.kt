package io.nekohasekai.sfa.ui.dashboard

import io.nekohasekai.libbox.Provider as LibboxProvider
import io.nekohasekai.libbox.ProviderItem as LibboxProviderItem
import io.nekohasekai.libbox.ProviderItemIterator

data class Provider(
    val tag: String,
    val type: String,
    val info: Map<String, Long>,
    val listUpdateTime: Long,
    var isExpand: Boolean,
    var items: List<ProviderItem>,
) {
    constructor(item: LibboxProvider) : this(
        item.tag,
        item.type,
        info = mapOf(
            "Download" to item.info.download,
            "Upload" to item.info.upload,
            "Total" to item.info.total,
            "Expire" to item.info.expire
        ),
        item.lastUpdateTime,
        item.isExpand,
        item.items.toList().map { ProviderItem(it) },
    )
}

data class ProviderItem(
    val tag: String,
    val type: String,
    val urlTestTime: Long,
    val urlTestDelay: Int,
) {
    constructor(item: LibboxProviderItem) : this(
        item.tag,
        item.type,
        item.urlTestTime,
        item.urlTestDelay,
    )
}

internal fun ProviderItemIterator.toList(): List<LibboxProviderItem> {
    val list = mutableListOf<LibboxProviderItem>()
    while (hasNext()) {
        list.add(next())
    }
    return list
}

package io.nekohasekai.sfa.ui.dashboard

import android.annotation.SuppressLint
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.core.view.isVisible
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.GridLayoutManager
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import androidx.recyclerview.widget.SimpleItemAnimator

import io.nekohasekai.sfa.databinding.FragmentDashboardProvidersBinding
import io.nekohasekai.sfa.databinding.ViewDashboardProviderBinding
import io.nekohasekai.sfa.databinding.ViewDashboardProviderItemBinding

import io.nekohasekai.sfa.R
import io.nekohasekai.libbox.Libbox
import io.nekohasekai.libbox.Provider as LibboxProvider
import io.nekohasekai.sfa.constant.Status
import io.nekohasekai.sfa.ui.MainActivity
import io.nekohasekai.sfa.utils.CommandClient
import io.nekohasekai.sfa.ktx.colorForURLTestDelay
import io.nekohasekai.sfa.ktx.errorDialogBuilder

import kotlinx.coroutines.DelicateCoroutinesApi
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.GlobalScope
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

class ProvidersFragment : Fragment(), CommandClient.Handler {

    private val activity: MainActivity? get() = super.getActivity() as MainActivity?
    private var binding: FragmentDashboardProvidersBinding? = null
    private var adapter: Adapter? = null
    private val commandClient =
        CommandClient(lifecycleScope, CommandClient.ConnectionType.Providers, this)

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?
    ): View {
        val binding = FragmentDashboardProvidersBinding.inflate(inflater, container, false)
        this.binding = binding
        onCreate()
        return binding.root
    }

    private fun onCreate() {
        val activity = activity ?: return
        val binding = binding ?: return
        adapter = Adapter()
        binding.container.adapter = adapter
        binding.container.layoutManager = LinearLayoutManager(requireContext())
        activity.serviceStatus.observe(viewLifecycleOwner) {
            if (it == Status.Started) {
                commandClient.connect()
            }
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        binding = null
        commandClient.disconnect()
    }
    
    private var displayed = false
    private fun updateDisplayed(newValue: Boolean) {
        val binding = binding ?: return
        if (displayed != newValue) {
            displayed = newValue
            binding.statusText.isVisible = !displayed
            binding.container.isVisible = displayed
        }
    }

    override fun onConnected() {
        lifecycleScope.launch(Dispatchers.Main) {
            updateDisplayed(true)
        }
    }

    override fun onDisconnected() {
        lifecycleScope.launch(Dispatchers.Main) {
            updateDisplayed(false)
        }
    }

    @SuppressLint("NotifyDataSetChanged")
    override fun updateProviders(newProviders: MutableList<LibboxProvider>) {
        val adapter = adapter ?: return
        activity?.runOnUiThread {
            updateDisplayed(newProviders.isNotEmpty())
            adapter.setProviders(newProviders.map(::Provider))
        }
    }

    private class Adapter : RecyclerView.Adapter<ProviderView>() {

        private lateinit var providers: MutableList<Provider>

        @SuppressLint("NotifyDataSetChanged")
        fun setProviders(newProviders: List<Provider>) {
            providers = newProviders.toMutableList()
            notifyDataSetChanged()
        }

        override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): ProviderView {
            return ProviderView(
                ViewDashboardProviderBinding.inflate(
                    LayoutInflater.from(parent.context),
                    parent,
                    false
                ),
            )
        }

        override fun getItemCount(): Int {
            if (!::providers.isInitialized) {
                return 0
            }
            return providers.size
        }

        override fun onBindViewHolder(holder: ProviderView, position: Int) {
            holder.bind(providers[position])
        }
    }

    private class ProviderView(val binding: ViewDashboardProviderBinding) :
        RecyclerView.ViewHolder(binding.root) {

        private lateinit var provider: Provider
        private lateinit var items: List<ProviderItem>
        private lateinit var adapter: ItemAdapter

        @OptIn(DelicateCoroutinesApi::class)
        @SuppressLint("NotifyDataSetChanged")
        fun bind(provider: Provider) {
            this.provider = provider
            binding.providerName.text = provider.tag
            binding.providerType.text = Libbox.providerDisplayType(provider.type)

            updateInfo()

            binding.providerUpdateButton.setOnClickListener {
                GlobalScope.launch {
                    runCatching {
                        Libbox.newStandaloneCommandClient().providerUpdate(provider.tag)
                    }.onFailure {
                        withContext(Dispatchers.Main) {
                            binding.root.context.errorDialogBuilder(it).show()
                        }
                    }
                }
            }
            items = provider.items
            if (!::adapter.isInitialized) {
                adapter = ItemAdapter(provider, items.toMutableList())
                binding.itemList.adapter = adapter
                (binding.itemList.itemAnimator as SimpleItemAnimator).supportsChangeAnimations =
                    false
                binding.itemList.layoutManager = GridLayoutManager(binding.root.context, 2)
            } else {
                adapter.provider = provider
                adapter.setItems(items)
            }
            updateExpand()
        }

        @SuppressLint("SetTextI18n")
        private fun updateInfo() {
            @SuppressLint("DefaultLocale")
            fun formatBytes(bytes: Long): String {
                val gib = bytes.toDouble() / (1024.0 * 1024 * 1024)
                return String.format("%.1f GiB", gib)
            }

            @SuppressLint("DefaultLocale")
            fun formatPercentage(used: Long, total: Long): String {
                if (total == 0L) return "0%"
                val percent = used.toDouble() / total * 100
                return String.format("%.2f%%", percent)
            }

            fun formatExpireTime(timestamp: Long): String {
                val sdf = SimpleDateFormat("yyyy-MM-dd", Locale.getDefault())
                return "到期时间: ${sdf.format(Date(timestamp * 1000))}"
            }

            fun formatUpdatedAt(secondsAgo: Long): String {
                val minutes = secondsAgo / 60
                val hours = minutes / 60
                return when {
                    hours > 0 -> "更新于 $hours 小时前"
                    minutes > 0 -> "更新于 $minutes 分钟前"
                    else -> "刚刚更新"
                }
            }

            val data = provider.info
            val download = data["Download"] ?: 0L
            val upload = data["Upload"] ?: 0L
            val expire = data["Expire"] ?: 0L
            val total = data["Total"] ?: 0L

            val totalUsed = download + upload
            binding.expirationTime.text = formatExpireTime(expire)
            binding.trafficData.text = "${formatBytes(totalUsed)} / ${formatBytes(total)} ( ${formatPercentage(totalUsed, total)} )"

            val now = System.currentTimeMillis() / 1000
            val secondsAgo = now - provider.listUpdateTime
            binding.updateAt.text = formatUpdatedAt(secondsAgo)
        }

        @OptIn(DelicateCoroutinesApi::class)
        private fun updateExpand(isExpand: Boolean? = null) {
            val newExpandStatus = isExpand ?: provider.isExpand
            if (isExpand != null) {
                GlobalScope.launch {
                    runCatching {
                        Libbox.newStandaloneCommandClient().setProviderExpand(provider.tag, isExpand)
                    }.onFailure {
                        withContext(Dispatchers.Main) {
                            binding.root.context.errorDialogBuilder(it).show()
                        }
                    }
                }
            }
            binding.itemList.isVisible = newExpandStatus
            if (newExpandStatus) {
                binding.expandButton.setImageResource(R.drawable.ic_expand_less_24)
            } else {
                binding.expandButton.setImageResource(R.drawable.ic_expand_more_24)
            }
            binding.expandButton.setOnClickListener {
                updateExpand(!binding.itemList.isVisible)
            }
        }
    }

    private class ItemAdapter(
        var provider: Provider,
        private var items: MutableList<ProviderItem> = mutableListOf()
    ) :
        RecyclerView.Adapter<ItemProviderView>() {

        @SuppressLint("NotifyDataSetChanged")
        fun setItems(newItems: List<ProviderItem>) {
            if (items.size != newItems.size) {
                items = newItems.toMutableList()
                notifyDataSetChanged()
            } else {
                newItems.forEachIndexed { index, item ->
                    if (items[index] != item) {
                        items[index] = item
                        notifyItemChanged(index)
                    }
                }
            }
        }

        override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): ItemProviderView {
            return ItemProviderView(
                ViewDashboardProviderItemBinding.inflate(
                    LayoutInflater.from(parent.context),
                    parent,
                    false
                )
            )
        }

        override fun getItemCount(): Int {
            return items.size
        }

        override fun onBindViewHolder(holder: ItemProviderView, position: Int) {
            holder.bind(items[position])
        }
    }

    private class ItemProviderView(val binding: ViewDashboardProviderItemBinding) :
        RecyclerView.ViewHolder(binding.root) {

        @SuppressLint("SetTextI18n")
        fun bind(item: ProviderItem) {
            binding.itemName.text = item.tag
            binding.itemType.text = Libbox.providerDisplayType(item.type)
            binding.itemStatus.isVisible = item.urlTestTime > 0
            if (item.urlTestTime > 0) {
                binding.itemStatus.text = "${item.urlTestDelay}ms"
                binding.itemStatus.setTextColor(
                    colorForURLTestDelay(
                        binding.root.context,
                        item.urlTestDelay
                    )
                )
            }
        }
    }
}